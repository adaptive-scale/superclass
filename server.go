package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/adaptive-scale/superclass/pkg/classifier"
	"github.com/adaptive-scale/superclass/pkg/extractor"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	uploadDir string
	provider  classifier.Provider
	config    classifier.ModelConfig
}

type ClassificationRequest struct {
	Categories []string `json:"categories,omitempty"`
}

type ClassificationResponse struct {
	Category   string   `json:"category"`
	Confidence float64  `json:"confidence"`
	Summary    string   `json:"summary"`
	Keywords   []string `json:"keywords"`
	RawText    string   `json:"raw_text,omitempty"`
	Error      string   `json:"error,omitempty"`
}

func NewServer(uploadDir string, provider classifier.Provider, config classifier.ModelConfig) *Server {
	return &Server{
		uploadDir: uploadDir,
		provider:  provider,
		config:    config,
	}
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvFloat64WithDefault gets a float64 environment variable with a default value
func getEnvFloat64WithDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvIntWithDefault gets an int environment variable with a default value
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func NewServerFromEnv() *Server {
	log.Debug("Starting server initialization from environment")

	// Get configuration from environment variables
	uploadDir := getEnvWithDefault("UPLOAD_DIR", "/tmp/superclass-uploads")
	modelType := getEnvWithDefault("MODEL_TYPE", "gpt-4")
	provider := classifier.ProviderFromString(getEnvWithDefault("MODEL_PROVIDER", "openai"))
	maxCost := getEnvFloat64WithDefault("MAX_COST", 0.1)
	maxLatency := getEnvIntWithDefault("MAX_LATENCY", 30)

	log.WithFields(log.Fields{
		"uploadDir":  uploadDir,
		"modelType":  modelType,
		"provider":   provider,
		"maxCost":    maxCost,
		"maxLatency": maxLatency,
	}).Info("Server configuration loaded")

	log.Debug("Creating model configuration")
	// Create model config
	config := classifier.ModelConfig{
		Model:  modelType,
		APIKey: os.Getenv("OPENAI_API_KEY"), // Will be overridden by provider-specific key
		Parameters: map[string]interface{}{
			"max_tokens":  2000,
			"temperature": 0.3,
			"max_cost":    maxCost,
			"max_latency": maxLatency,
		},
	}

	log.Debug("Setting provider-specific API key")
	// Set the appropriate API key based on the provider
	switch provider {
	case classifier.OpenAI:
		config.APIKey = os.Getenv("OPENAI_API_KEY")
		log.Debug("Using OpenAI provider")
	case classifier.Anthropic:
		config.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		log.Debug("Using Anthropic provider")
	case classifier.Azure:
		config.APIKey = os.Getenv("AZURE_OPENAI_API_KEY")
		log.Debug("Using Azure OpenAI provider")
	}

	log.Debug("Server initialization completed")
	return NewServer(uploadDir, provider, config)
}

func (s *Server) handleClassify(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"handler": "classify",
		"method":  r.Method,
		"remote":  r.RemoteAddr,
	})

	if r.Method != http.MethodPost {
		logger.Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with 32MB max memory
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		logger.WithError(err).Error("Failed to parse form")
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Parse optional categories from form
	var classificationReq ClassificationRequest
	if categoriesJSON := r.FormValue("categories"); categoriesJSON != "" {
		if err := json.Unmarshal([]byte(categoriesJSON), &classificationReq.Categories); err != nil {
			logger.WithError(err).Error("Failed to parse categories")
			http.Error(w, "Invalid categories format", http.StatusBadRequest)
			return
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logger.WithError(err).Error("Failed to get file from form")
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	logger = logger.WithFields(log.Fields{
		"filename":       header.Filename,
		"size":           header.Size,
		"content_type":   header.Header.Get("Content-Type"),
		"has_categories": len(classificationReq.Categories) > 0,
	})
	logger.Info("Processing uploaded file")

	// Create temporary file
	tempFile := filepath.Join(s.uploadDir, header.Filename)
	out, err := os.Create(tempFile)
	if err != nil {
		logger.WithError(err).Error("Failed to create temporary file")
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer func() {
		out.Close()
		if err := os.Remove(tempFile); err != nil {
			logger.WithError(err).Warn("Failed to remove temporary file")
		}
	}()

	// Copy uploaded file to temporary file
	if _, err := io.Copy(out, file); err != nil {
		logger.WithError(err).Error("Failed to save file")
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	logger.Debug("Starting classification")
	// Extract and classify
	result, err := extractor.ExtractAndClassifyWithOptions(tempFile, s.provider, s.config, classifier.ClassificationOptions{
		Categories: classificationReq.Categories,
	})
	if err != nil {
		logger.WithError(err).Error("Classification failed")
		json.NewEncoder(w).Encode(ClassificationResponse{
			Error: err.Error(),
		})
		return
	}

	// Prepare response
	response := ClassificationResponse{
		Category:   result.Classification.Category,
		Confidence: result.Classification.Confidence,
		Summary:    result.Classification.Summary,
		Keywords:   result.Classification.Keywords,
		RawText:    result.Text,
	}

	logger.WithFields(log.Fields{
		"category":        response.Category,
		"confidence":      response.Confidence,
		"keywords":        response.Keywords,
		"used_categories": classificationReq.Categories,
	}).Info("Classification completed successfully")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"handler":    "health",
		"method":     r.Method,
		"remote":     r.RemoteAddr,
		"request_id": r.Header.Get("X-Request-ID"),
	})

	logger.Debug("Health check requested")

	// Add basic system metrics in debug mode
	logger.WithFields(log.Fields{
		"uptime":     time.Since(startTime).String(),
		"goroutines": runtime.NumGoroutine(),
		"cpu_cores":  runtime.NumCPU(),
	}).Debug("System metrics")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	logger.Debug("Health check completed")
}

var startTime time.Time

func (s *Server) Start(port int) error {
	startTime = time.Now()

	log.Debug("Ensuring upload directory exists")
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		log.WithError(err).Error("Failed to create upload directory")
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	log.Debug("Registering HTTP handlers")
	// Register routes
	http.HandleFunc("/classify", s.handleClassify)
	http.HandleFunc("/health", s.handleHealth)

	// Start server
	addr := fmt.Sprintf(":%d", port)
	log.WithFields(log.Fields{
		"address":    addr,
		"port":       port,
		"upload_dir": s.uploadDir,
		"provider":   s.provider,
		"model":      s.config.Model,
	}).Info("Server starting on port %d", port)

	log.Debug("Starting HTTP server")
	return http.ListenAndServe(addr, nil)
}
