package extractor

import (
	"encoding/json"
	"fmt"

	"github.com/adaptive-scale/superclass/pkg/classifier"
	log "github.com/sirupsen/logrus"
)

// DocumentFeatures represents various features extracted from a document
type DocumentFeatures struct {
	// Basic statistics
	WordCount        int     `json:"word_count"`
	CharCount        int     `json:"char_count"`
	SentenceCount    int     `json:"sentence_count"`
	AverageWordLen   float64 `json:"avg_word_length"`
	UniqueWordCount  int     `json:"unique_word_count"`
	ParagraphCount   int     `json:"paragraph_count"`

	// Language features
	TopKeywords      []string          `json:"top_keywords"`
	NamedEntities    []NamedEntity     `json:"named_entities"`
	SentimentScore   float64           `json:"sentiment_score"`
	LanguageMetrics  LanguageMetrics   `json:"language_metrics"`
	ContentStructure ContentStructure  `json:"content_structure"`
}

// NamedEntity represents an entity detected in the text
type NamedEntity struct {
	Text  string `json:"text"`
	Label string `json:"label"` // PERSON, ORGANIZATION, LOCATION, etc.
}

// LanguageMetrics contains various language-related metrics
type LanguageMetrics struct {
	ReadabilityScore   float64 `json:"readability_score"`
	TechnicalityScore  float64 `json:"technicality_score"`
	FormalityScore     float64 `json:"formality_score"`
	VocabularyRichness float64 `json:"vocabulary_richness"`
}

// ContentStructure represents the document's structural elements
type ContentStructure struct {
	HeadingCount     int      `json:"heading_count"`
	ListCount        int      `json:"list_count"`
	TableCount       int      `json:"table_count"`
	CodeBlockCount   int      `json:"code_block_count"`
	ImageCount       int      `json:"image_count"`
	HeadingHierarchy []string `json:"heading_hierarchy"`
}

// ModelPrompts contains feature extraction prompts for different models
var ModelPrompts = map[classifier.Provider]string{
	classifier.OpenAI: `You are a document analysis expert. Analyze the following text and extract key features. Return ONLY a JSON object with this exact structure:
{
  "word_count": int,
  "char_count": int,
  "sentence_count": int,
  "avg_word_length": float,
  "unique_word_count": int,
  "paragraph_count": int,
  "top_keywords": [string],
  "named_entities": [{"text": string, "label": string}],
  "sentiment_score": float,
  "language_metrics": {
    "readability_score": float,
    "technicality_score": float,
    "formality_score": float,
    "vocabulary_richness": float
  },
  "content_structure": {
    "heading_count": int,
    "list_count": int,
    "table_count": int,
    "code_block_count": int,
    "image_count": int,
    "heading_hierarchy": [string]
  }
}

Consider:
1. Basic text statistics
2. Key topics and themes
3. Named entities (people, organizations, locations)
4. Document structure and formatting
5. Language complexity and style
6. Technical vs non-technical content
7. Formal vs informal language
8. Sentiment analysis (score from -1.0 to 1.0)

Text to analyze:`,

	classifier.Anthropic: `You are Claude, a document analysis expert. Your task is to analyze the provided text and return a JSON object containing detailed features. The response must be ONLY the JSON object, no other text.

Required JSON structure:
{
  "word_count": int,
  "char_count": int,
  "sentence_count": int,
  "avg_word_length": float,
  "unique_word_count": int,
  "paragraph_count": int,
  "top_keywords": [string],
  "named_entities": [{"text": string, "label": string}],
  "sentiment_score": float,
  "language_metrics": {
    "readability_score": float,
    "technicality_score": float,
    "formality_score": float,
    "vocabulary_richness": float
  },
  "content_structure": {
    "heading_count": int,
    "list_count": int,
    "table_count": int,
    "code_block_count": int,
    "image_count": int,
    "heading_hierarchy": [string]
  }
}

Analyze for:
1. Basic text statistics (counts, lengths)
2. Key topics and themes (keywords)
3. Named entities (people, organizations, locations)
4. Document structure (headings, lists, code blocks)
5. Language complexity and style metrics
6. Technical content indicators
7. Formality assessment
8. Sentiment (score from -1.0 to 1.0)

Text to analyze:`,

	classifier.Azure: `You are an AI language model specializing in document analysis. Extract features from the provided text and return them in a specific JSON format. Return ONLY the JSON object, no other text.

Required JSON structure:
{
  "word_count": int,
  "char_count": int,
  "sentence_count": int,
  "avg_word_length": float,
  "unique_word_count": int,
  "paragraph_count": int,
  "top_keywords": [string],
  "named_entities": [{"text": string, "label": string}],
  "sentiment_score": float,
  "language_metrics": {
    "readability_score": float,
    "technicality_score": float,
    "formality_score": float,
    "vocabulary_richness": float
  },
  "content_structure": {
    "heading_count": int,
    "list_count": int,
    "table_count": int,
    "code_block_count": int,
    "image_count": int,
    "heading_hierarchy": [string]
  }
}

Analysis criteria:
1. Text statistics (word, character, sentence counts)
2. Keywords and themes
3. Named entity recognition
4. Document structure analysis
5. Language complexity metrics
6. Technical content assessment
7. Formality level
8. Sentiment analysis (-1.0 to 1.0)

Text to analyze:`,
}

// DefaultModelConfig returns the recommended model configuration for feature extraction
func DefaultModelConfig(provider classifier.Provider) classifier.ModelConfig {
	switch provider {
	case classifier.OpenAI:
		return classifier.ModelConfig{
			Model: string(classifier.GPT4Turbo),
			Parameters: map[string]interface{}{
				"temperature": 0.1, // Low temperature for consistent analysis
				"max_tokens":  2000,
			},
		}
	case classifier.Anthropic:
		return classifier.ModelConfig{
			Model: string(classifier.Claude3Opus),
			Parameters: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  2000,
			},
		}
	case classifier.Azure:
		return classifier.ModelConfig{
			Model: string(classifier.AzureGPT4),
			Parameters: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  2000,
			},
		}
	default:
		return classifier.ModelConfig{
			Model: string(classifier.GPT4Turbo),
			Parameters: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  2000,
			},
		}
	}
}

// ExtractFeatures extracts various features from the document text using the specified model
func ExtractFeatures(text string, provider classifier.Provider, config classifier.ModelConfig) (*DocumentFeatures, error) {
	logger := log.WithFields(log.Fields{
		"function": "ExtractFeatures",
		"provider": provider,
		"model":    config.Model,
	})
	logger.Debug("Starting model-based feature extraction")

	// Create classifier for the specified provider
	clf, err := classifier.NewClassifier(provider, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create classifier: %w", err)
	}

	// Get the appropriate prompt for the provider
	prompt, ok := ModelPrompts[provider]
	if !ok {
		prompt = ModelPrompts[classifier.OpenAI] // Fallback to OpenAI prompt
	}

	// Get model's analysis
	response, err := clf.Classify(prompt + "\n\n" + text)
	if err != nil {
		return nil, fmt.Errorf("model analysis failed: %w", err)
	}

	// Parse model's JSON response
	var features DocumentFeatures
	if err := json.Unmarshal([]byte(response.Category), &features); err != nil {
		return nil, fmt.Errorf("failed to parse model response: %w", err)
	}

	logger.WithFields(log.Fields{
		"word_count":     features.WordCount,
		"sentence_count": features.SentenceCount,
		"entity_count":   len(features.NamedEntities),
	}).Debug("Feature extraction completed")

	return &features, nil
}

// ExtractFeaturesAndClassify extracts features and classifies the document
func ExtractFeaturesAndClassify(path string, provider classifier.Provider, config classifier.ModelConfig) (*ExtractResult, *DocumentFeatures, error) {
	logger := log.WithFields(log.Fields{
		"function": "ExtractFeaturesAndClassify",
		"path":     path,
		"provider": provider,
	})
	logger.Debug("Starting feature extraction and classification")

	// First extract and classify the text
	result, err := ExtractAndClassify(path, provider, config)
	if err != nil {
		return nil, nil, err
	}

	// Then extract features using the same provider
	features, err := ExtractFeatures(result.Text, provider, config)
	if err != nil {
		return result, nil, err
	}

	return result, features, nil
} 