package classifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// AnthropicClassifier handles content classification using Anthropic's Claude models
type AnthropicClassifier struct {
	apiKey     string
	model      string
	endpoint   string
	parameters map[string]interface{}
}

// NewAnthropicClassifier creates a new Anthropic classifier
func NewAnthropicClassifier(config ModelConfig) *AnthropicClassifier {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1/messages"
	}

	model := config.Model
	if model == "" {
		model = "claude-3-opus-20240229" // Default to latest Claude model
	}

	return &AnthropicClassifier{
		apiKey:     config.APIKey,
		model:      model,
		endpoint:   endpoint,
		parameters: config.Parameters,
	}
}

// Configure updates the classifier configuration
func (c *AnthropicClassifier) Configure(config ModelConfig) error {
	if config.APIKey != "" {
		c.apiKey = config.APIKey
	}
	if config.Endpoint != "" {
		c.endpoint = config.Endpoint
	}
	if config.Model != "" {
		c.model = config.Model
	}
	if config.Parameters != nil {
		c.parameters = config.Parameters
	}
	return nil
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Classify takes text content and returns classification details
func (c *AnthropicClassifier) Classify(content string) (*Classification, error) {
	return c.ClassifyWithOptions(content, ClassificationOptions{})
}

// ClassifyWithOptions takes text content and classification options and returns classification details
func (c *AnthropicClassifier) ClassifyWithOptions(content string, options ClassificationOptions) (*Classification, error) {
	logger := logrus.WithFields(logrus.Fields{
		"function":       "ClassifyWithOptions",
		"model":          c.model,
		"content_length": len(content),
		"has_categories": len(options.Categories) > 0,
	})
	logger.Debug("Starting content classification")

	if c.apiKey == "" {
		logger.Error("Anthropic API key is required")
		return nil, fmt.Errorf("Anthropic API key is required")
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

	reqBody := anthropicRequest{
		Model: c.model,
		Messages: []anthropicMessage{
			{
				Role:    "system",
				Content: "You are a content classification expert. Always respond in valid JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: c.parameters["max_tokens"].(int),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal request")
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"request_body":          string(jsonBody),
		"model":                 c.model,
		"predefined_categories": options.Categories,
	}).Debug("Request payload prepared")

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.WithError(err).Error("Failed to create request")
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Error("API request failed")
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.WithField("status_code", resp.StatusCode).Error("API request failed")
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		logger.WithError(err).Error("Failed to decode response")
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		logger.Error("No classification result received")
		return nil, fmt.Errorf("no classification result received")
	}

	var classification Classification
	if err := json.Unmarshal([]byte(anthropicResp.Content[0].Text), &classification); err != nil {
		logger.WithFields(logrus.Fields{
			"raw_content": anthropicResp.Content[0].Text,
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
			logger.WithFields(logrus.Fields{
				"received_category": classification.Category,
				"valid_categories":  options.Categories,
			}).Error("Classification returned invalid category")
			return nil, fmt.Errorf("classifier returned invalid category: %s", classification.Category)
		}
	}

	logger.WithFields(logrus.Fields{
		"category":                   classification.Category,
		"confidence":                 classification.Confidence,
		"keywords_count":             len(classification.Keywords),
		"summary_length":             len(classification.Summary),
		"used_predefined_categories": len(options.Categories) > 0,
	}).Debug("Classification completed successfully")

	return &classification, nil
}
