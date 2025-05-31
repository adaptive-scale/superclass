package classifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CustomClassifier is a template for implementing custom model classifiers
type CustomClassifier struct {
	apiKey     string
	model      string
	endpoint   string
	parameters map[string]interface{}
}

// customMessage represents a message in the custom API request
type customMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// customRequest represents the request structure for the custom API
type customRequest struct {
	Model      string                 `json:"model"`
	Messages   []customMessage        `json:"messages"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// customResponse represents the response structure from the custom API
type customResponse struct {
	Content string `json:"content"`
}

// NewCustomClassifier creates a new custom classifier instance
func NewCustomClassifier(config ModelConfig) *CustomClassifier {
	return &CustomClassifier{
		apiKey:     config.APIKey,
		model:      config.Model,
		endpoint:   config.Endpoint,
		parameters: config.Parameters,
	}
}

// Configure updates the classifier configuration
func (c *CustomClassifier) Configure(config ModelConfig) error {
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

// Classify takes text content and returns classification details
func (c *CustomClassifier) Classify(content string) (*Classification, error) {
	return c.ClassifyWithOptions(content, ClassificationOptions{})
}

// ClassifyWithOptions takes text content and classification options and returns classification details
func (c *CustomClassifier) ClassifyWithOptions(content string, options ClassificationOptions) (*Classification, error) {
	logger := log.WithFields(log.Fields{
		"function":       "ClassifyWithOptions",
		"model":          c.model,
		"content_length": len(content),
		"has_categories": len(options.Categories) > 0,
	})
	logger.Debug("Starting content classification")

	if c.endpoint == "" {
		logger.Error("Custom endpoint URL is required")
		return nil, fmt.Errorf("Custom endpoint URL is required")
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

	reqBody := customRequest{
		Model: c.model,
		Messages: []customMessage{
			{
				Role:    "system",
				Content: "You are a content classification expert. Always respond in valid JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Parameters: c.parameters,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal request")
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	logger.WithFields(log.Fields{
		"request_body":          string(jsonBody),
		"model":                 c.model,
		"predefined_categories": options.Categories,
	}).Debug("Request payload prepared")

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.WithError(err).Error("Failed to create request")
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

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

	var customResp customResponse
	if err := json.NewDecoder(resp.Body).Decode(&customResp); err != nil {
		logger.WithError(err).Error("Failed to decode response")
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	var classification Classification
	if err := json.Unmarshal([]byte(customResp.Content), &classification); err != nil {
		logger.WithFields(log.Fields{
			"raw_content": customResp.Content,
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
		"used_predefined_categories": len(options.Categories) > 0,
	}).Debug("Classification completed successfully")

	return &classification, nil
}

/* Example implementation:

func (c *CustomClassifier) Classify(content string) (*Classification, error) {
	// 1. Validate configuration
	if c.apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// 2. Prepare the request
	// - Format your API request body
	// - Set up headers
	// - Configure any model-specific parameters

	// 3. Make the API call
	// - Use net/http or your preferred HTTP client
	// - Handle response status codes
	// - Parse the response

	// 4. Process the results
	// - Extract relevant information
	// - Map to Classification struct
	// - Handle errors appropriately

	// 5. Return the results
	return &Classification{
		Category:   "your-category",
		Confidence: 0.95,
		Summary:    "your-summary",
		Keywords:   []string{"keyword1", "keyword2"},
	}, nil
}
*/
