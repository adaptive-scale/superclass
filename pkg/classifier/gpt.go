package classifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GPTClassifier handles content classification using OpenAI's GPT models
type GPTClassifier struct {
	apiKey     string
	model      string
	endpoint   string
	parameters map[string]interface{}
}

// NewGPTClassifier creates a new GPT classifier
func NewGPTClassifier(config ModelConfig) *GPTClassifier {
	logger := log.WithFields(log.Fields{
		"function": "NewGPTClassifier",
		"model":    config.Model,
	})
	logger.Debug("Creating new GPT classifier")

	apiKey := config.APIKey
	if apiKey == "" {
		logger.Debug("API key not provided in config, checking environment")
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	endpoint := config.Endpoint
	if endpoint == "" {
		logger.Debug("Using default OpenAI endpoint")
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	model := config.Model
	if model == "" {
		logger.Debug("Using default GPT model")
		model = "gpt-3.5-turbo"
	}

	logger.WithFields(log.Fields{
		"endpoint":     endpoint,
		"has_api_key":  apiKey != "",
		"model":        model,
		"params_count": len(config.Parameters),
	}).Debug("GPT classifier initialized")

	return &GPTClassifier{
		apiKey:     apiKey,
		model:      model,
		endpoint:   endpoint,
		parameters: config.Parameters,
	}
}

// Configure updates the classifier configuration
func (c *GPTClassifier) Configure(config ModelConfig) error {
	logger := log.WithFields(log.Fields{
		"function":      "Configure",
		"current_model": c.model,
		"new_model":     config.Model,
	})
	logger.Debug("Updating GPT classifier configuration")

	if config.APIKey != "" {
		logger.Debug("Updating API key")
		c.apiKey = config.APIKey
	}
	if config.Endpoint != "" {
		logger.WithField("new_endpoint", config.Endpoint).Debug("Updating endpoint")
		c.endpoint = config.Endpoint
	}
	if config.Model != "" {
		logger.WithField("new_model", config.Model).Debug("Updating model")
		c.model = config.Model
	}
	if config.Parameters != nil {
		logger.WithField("params_count", len(config.Parameters)).Debug("Updating parameters")
		c.parameters = config.Parameters
	}

	logger.Debug("Configuration updated successfully")
	return nil
}

type gptRequest struct {
	Model       string       `json:"model"`
	Messages    []gptMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
}

type gptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type gptResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Classify takes text content and returns classification details
func (c *GPTClassifier) Classify(content string) (*Classification, error) {
	return c.ClassifyWithOptions(content, ClassificationOptions{})
}

// ClassifyWithOptions takes text content and classification options and returns classification details
func (c *GPTClassifier) ClassifyWithOptions(content string, options ClassificationOptions) (*Classification, error) {
	logger := log.WithFields(log.Fields{
		"function":       "ClassifyWithOptions",
		"model":          c.model,
		"content_length": len(content),
		"endpoint":       c.endpoint,
		"has_categories": len(options.Categories) > 0,
	})
	logger.Debug("Starting content classification")

	if c.apiKey == "" {
		logger.Error("Missing API key")
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	var prompt string
	if len(options.Categories) > 0 {
		categoriesStr := strings.Join(options.Categories, ", ")
		prompt = fmt.Sprintf(`Analyze the following text and classify it into one of these categories: %s

Provide a JSON response with these fields:
	- category: One of the categories listed above that best matches the content
	- confidence: A confidence score between 0 and 1 indicating how well the content matches the chosen category
	- summary: A brief summary of the content (max 100 words)
	- keywords: Up to 5 key terms or phrases from the content

Text to analyze:
%s`, categoriesStr, content)
	} else {
		prompt = fmt.Sprintf(`Analyze the following text and provide a JSON response with these fields:
	- category: The main category/topic of the content
	- confidence: A confidence score between 0 and 1
	- summary: A brief summary of the content (max 100 words)
	- keywords: Up to 5 key terms or phrases from the content

Text to analyze:
%s`, content)
	}

	// Extract parameters from the config
	temperature := 0.3 // default temperature
	maxTokens := 2000  // default max tokens
	if c.parameters != nil {
		if temp, ok := c.parameters["temperature"].(float64); ok {
			temperature = temp
		}
		if tokens, ok := c.parameters["max_tokens"].(int); ok {
			maxTokens = tokens
		}
	}

	logger.Debug("Preparing API request")
	reqBody := gptRequest{
		Model: c.model,
		Messages: []gptMessage{
			{
				Role:    "system",
				Content: "You are a content classification expert. Always respond in valid JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal request")
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	logger.WithFields(log.Fields{
		"request_body":          string(jsonBody),
		"model":                 c.model,
		"temperature":           temperature,
		"max_tokens":            maxTokens,
		"predefined_categories": options.Categories,
	}).Debug("Request payload prepared")

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.WithError(err).Error("Failed to create request")
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	logger.Debug("Sending request to OpenAI API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Error("API request failed")
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.WithError(err).Error("Failed to read response body")
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var gptResp gptResponse
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&gptResp); err != nil {
		logger.WithError(err).Error("Failed to decode response")
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errorMsg := "unknown error"
		if gptResp.Error.Message != "" {
			errorMsg = gptResp.Error.Message
		}

		logger.WithFields(log.Fields{
			"status_code":   resp.StatusCode,
			"error_type":    gptResp.Error.Type,
			"error_code":    gptResp.Error.Code,
			"error_message": errorMsg,
			"request_id":    resp.Header.Get("X-Request-Id"),
			"request_body":  string(jsonBody),
		}).Error("API request failed")

		return nil, fmt.Errorf("API request failed: %s (type: %s, code: %s)",
			errorMsg, gptResp.Error.Type, gptResp.Error.Code)
	}

	if len(gptResp.Choices) == 0 {
		logger.Error("No classification result received")
		return nil, fmt.Errorf("no classification result received")
	}

	logger.Debug("Parsing classification result")
	var classification Classification
	if err := json.Unmarshal([]byte(gptResp.Choices[0].Message.Content), &classification); err != nil {
		logger.WithFields(log.Fields{
			"raw_content": gptResp.Choices[0].Message.Content,
		}).WithError(err).Error("Failed to parse classification")
		return nil, fmt.Errorf("error parsing classification: %w", err)
	}

	// Validate category if predefined categories were provided
	if len(options.Categories) > 0 {
		categoryValid := false
		for _, validCategory := range options.Categories {
			if strings.EqualFold(classification.Category, validCategory) {
				classification.Category = validCategory // Use exact case from predefined list
				categoryValid = true
				break
			}
		}
		if !categoryValid {
			logger.WithFields(log.Fields{
				"received_category": classification.Category,
				"valid_categories":  options.Categories,
			}).Error("Classification returned invalid category")
			return nil, fmt.Errorf("classifier returned invalid category: %s", classification.Category)
		}
	}

	logger.WithFields(log.Fields{
		"category":                   classification.Category,
		"confidence":                 classification.Confidence,
		"keywords_count":             len(classification.Keywords),
		"summary_length":             len(classification.Summary),
		"request_id":                 resp.Header.Get("X-Request-Id"),
		"used_predefined_categories": len(options.Categories) > 0,
	}).Debug("Classification completed successfully")

	return &classification, nil
}
