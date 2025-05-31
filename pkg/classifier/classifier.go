package classifier

import (
	log "github.com/sirupsen/logrus"
)

// Classification represents the classification result
type Classification struct {
	Category   string   `json:"category"`
	Confidence float64  `json:"confidence"`
	Summary    string   `json:"summary"`
	Keywords   []string `json:"keywords"`
}

// ModelConfig contains configuration for the AI model
type ModelConfig struct {
	// Endpoint URL for the model API
	Endpoint string
	// Model identifier/name
	Model string
	// API key or authentication token
	APIKey string
	// Additional model-specific parameters
	Parameters map[string]interface{}
	// Optional list of predefined categories to classify into
	PredefinedCategories []string
}

// ClassificationOptions contains options for classification
type ClassificationOptions struct {
	// List of categories to classify into. If empty, classifier will determine category freely.
	Categories []string
}

// Classifier defines the interface that all model classifiers must implement
type Classifier interface {
	// Classify analyzes the text content and returns classification details
	Classify(content string) (*Classification, error)
	// ClassifyWithOptions analyzes the text content with specific options
	ClassifyWithOptions(content string, options ClassificationOptions) (*Classification, error)
	// Configure updates the classifier configuration
	Configure(config ModelConfig) error
}

// Provider represents different AI model providers
type Provider string

const (
	OpenAI    Provider = "openai"
	Azure     Provider = "azure"
	Anthropic Provider = "anthropic"
	Custom    Provider = "custom"
)

// NewClassifier creates a new classifier instance for the specified provider
func NewClassifier(provider Provider, config ModelConfig) (Classifier, error) {
	logger := log.WithFields(log.Fields{
		"function": "NewClassifier",
		"provider": provider,
		"model":    config.Model,
	})
	logger.Debug("Creating new classifier instance")

	var classifier Classifier
	switch provider {
	case OpenAI:
		logger.Debug("Creating OpenAI GPT classifier")
		classifier = NewGPTClassifier(config)
	case Azure:
		logger.Debug("Creating Azure OpenAI classifier")
		classifier = NewAzureClassifier(config)
	case Anthropic:
		logger.Debug("Creating Anthropic Claude classifier")
		classifier = NewAnthropicClassifier(config)
	case Custom:
		logger.Debug("Creating custom classifier")
		classifier = NewCustomClassifier(config)
	default:
		logger.Debug("Using default OpenAI GPT classifier")
		classifier = NewGPTClassifier(config)
	}

	logger.WithFields(log.Fields{
		"endpoint":     config.Endpoint,
		"has_api_key":  config.APIKey != "",
		"params_count": len(config.Parameters),
	}).Debug("Classifier created successfully")

	return classifier, nil
}
