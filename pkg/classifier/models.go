package classifier

import (
	"fmt"
)

// ModelType represents a specific model from a provider
type ModelType string

// Predefined models for each provider
const (
	// OpenAI Models
	GPT4               ModelType = "gpt-4"
	GPT4Turbo          ModelType = "gpt-4-turbo-preview"
	GPT35Turbo         ModelType = "gpt-3.5-turbo"
	GPT35TurboInstruct ModelType = "gpt-3.5-turbo-instruct"

	// Anthropic Models
	Claude3Opus   ModelType = "claude-3-opus-20240229"
	Claude3Sonnet ModelType = "claude-3-sonnet-20240229"
	Claude3Haiku  ModelType = "claude-3-haiku-20240229"
	Claude2       ModelType = "claude-2.1"

	// Azure OpenAI Models (base names, deployment names are configured separately)
	AzureGPT4       ModelType = "gpt-4"
	AzureGPT35Turbo ModelType = "gpt-35-turbo"
)

// ModelCapability represents what a model is good at
type ModelCapability int

const (
	GeneralPurpose ModelCapability = iota
	HighAccuracy
	FastResponse
	LongContext
	CodeAnalysis
	MultilingualSupport
	StructuredOutput
	SemanticAnalysis
)

// ContentType represents different types of content for model recommendations
type ContentType int

const (
	GeneralText ContentType = iota
	TechnicalDoc
	CreativeWriting
	CodeSnippet
	LegalDocument
	AcademicPaper
	BusinessReport
	SocialMediaContent
)

// CostInfo contains pricing information for a model
type CostInfo struct {
	InputPerThousandTokens  float64 // Cost per 1K input tokens in USD
	OutputPerThousandTokens float64 // Cost per 1K output tokens in USD
	BatchProcessingSupport  bool    // Whether the model supports batch processing
	ConcurrentRequests      int     // Maximum concurrent requests allowed
}

// ModelInfo contains metadata about a model
type ModelInfo struct {
	Type         ModelType
	Provider     Provider
	Capabilities []ModelCapability
	MaxTokens    int
	Description  string
	Parameters   map[string]interface{}
	Cost         CostInfo
	AvgLatencyMs int // Average latency in milliseconds
}

// ModelCosts defines pricing for different models
var ModelCosts = map[ModelType]CostInfo{
	GPT4: {
		InputPerThousandTokens:  0.03,
		OutputPerThousandTokens: 0.06,
		BatchProcessingSupport:  true,
		ConcurrentRequests:      3500,
	},
	GPT4Turbo: {
		InputPerThousandTokens:  0.01,
		OutputPerThousandTokens: 0.03,
		BatchProcessingSupport:  true,
		ConcurrentRequests:      4000,
	},
	GPT35Turbo: {
		InputPerThousandTokens:  0.001,
		OutputPerThousandTokens: 0.002,
		BatchProcessingSupport:  true,
		ConcurrentRequests:      5000,
	},
	Claude3Opus: {
		InputPerThousandTokens:  0.015,
		OutputPerThousandTokens: 0.075,
		BatchProcessingSupport:  true,
		ConcurrentRequests:      4000,
	},
	Claude3Sonnet: {
		InputPerThousandTokens:  0.003,
		OutputPerThousandTokens: 0.015,
		BatchProcessingSupport:  true,
		ConcurrentRequests:      5000,
	},
}

// DefaultModelParams returns recommended parameters for each model
var DefaultModelParams = map[ModelType]map[string]interface{}{
	GPT4: {
		"temperature":       0.7,
		"max_tokens":        2000,
		"top_p":             1.0,
		"frequency_penalty": 0.0,
	},
	GPT35Turbo: {
		"temperature":       0.7,
		"max_tokens":        1000,
		"top_p":             1.0,
		"frequency_penalty": 0.0,
	},
	Claude3Opus: {
		"temperature": 0.7,
		"max_tokens":  3000,
		"top_k":       10,
		"top_p":       0.8,
	},
	Claude3Sonnet: {
		"temperature": 0.7,
		"max_tokens":  2000,
		"top_k":       10,
		"top_p":       0.8,
	},
}

// ModelRegistry contains information about available models
var ModelRegistry = map[ModelType]ModelInfo{
	// OpenAI Models
	GPT4: {
		Type:     GPT4,
		Provider: OpenAI,
		Capabilities: []ModelCapability{
			HighAccuracy,
			CodeAnalysis,
			LongContext,
			StructuredOutput,
			MultilingualSupport,
		},
		MaxTokens:    8192,
		Description:  "Most capable GPT-4 model, best for complex tasks requiring deep understanding",
		Parameters:   DefaultModelParams[GPT4],
		Cost:         ModelCosts[GPT4],
		AvgLatencyMs: 2000,
	},
	GPT35Turbo: {
		Type:     GPT35Turbo,
		Provider: OpenAI,
		Capabilities: []ModelCapability{
			GeneralPurpose,
			FastResponse,
			MultilingualSupport,
		},
		MaxTokens:    4096,
		Description:  "Fast and cost-effective model, good for most classification tasks",
		Parameters:   DefaultModelParams[GPT35Turbo],
		Cost:         ModelCosts[GPT35Turbo],
		AvgLatencyMs: 800,
	},

	// Anthropic Models
	Claude3Opus: {
		Type:     Claude3Opus,
		Provider: Anthropic,
		Capabilities: []ModelCapability{
			HighAccuracy,
			LongContext,
			CodeAnalysis,
			SemanticAnalysis,
			StructuredOutput,
		},
		MaxTokens:    100000,
		Description:  "Most capable Claude model, excellent for detailed analysis and classification",
		Parameters:   DefaultModelParams[Claude3Opus],
		Cost:         ModelCosts[Claude3Opus],
		AvgLatencyMs: 2500,
	},
	Claude3Sonnet: {
		Type:     Claude3Sonnet,
		Provider: Anthropic,
		Capabilities: []ModelCapability{
			GeneralPurpose,
			FastResponse,
			SemanticAnalysis,
		},
		MaxTokens:    50000,
		Description:  "Balanced Claude model, good performance and speed",
		Parameters:   DefaultModelParams[Claude3Sonnet],
		Cost:         ModelCosts[Claude3Sonnet],
		AvgLatencyMs: 1000,
	},
}

// EstimateCost calculates the estimated cost for processing text with a specific model
func EstimateCost(model ModelType, inputTokens, outputTokens int) float64 {
	info, exists := ModelRegistry[model]
	if !exists {
		return 0
	}

	inputCost := float64(inputTokens) * info.Cost.InputPerThousandTokens / 1000
	outputCost := float64(outputTokens) * info.Cost.OutputPerThousandTokens / 1000
	return inputCost + outputCost
}

// CompareModels compares two models and returns their differences
type ModelComparison struct {
	CostDiff           float64           // Difference in cost per 1K tokens
	LatencyDiff        int               // Difference in average latency
	SharedCapabilities []ModelCapability // Capabilities both models have
	UniqueToFirst      []ModelCapability // Capabilities only the first model has
	UniqueToSecond     []ModelCapability // Capabilities only the second model has
	TokenLimitDiff     int               // Difference in token limits
}

func CompareModels(model1, model2 ModelType) (*ModelComparison, error) {
	info1, exists1 := ModelRegistry[model1]
	info2, exists2 := ModelRegistry[model2]
	if !exists1 || !exists2 {
		return nil, fmt.Errorf("one or both models not found")
	}

	// Calculate cost difference (using input cost as baseline)
	costDiff := info1.Cost.InputPerThousandTokens - info2.Cost.InputPerThousandTokens

	// Find shared and unique capabilities
	capMap1 := make(map[ModelCapability]bool)
	capMap2 := make(map[ModelCapability]bool)
	for _, cap := range info1.Capabilities {
		capMap1[cap] = true
	}
	for _, cap := range info2.Capabilities {
		capMap2[cap] = true
	}

	var shared, unique1, unique2 []ModelCapability
	for cap := range capMap1 {
		if capMap2[cap] {
			shared = append(shared, cap)
		} else {
			unique1 = append(unique1, cap)
		}
	}
	for cap := range capMap2 {
		if !capMap1[cap] {
			unique2 = append(unique2, cap)
		}
	}

	return &ModelComparison{
		CostDiff:           costDiff,
		LatencyDiff:        info1.AvgLatencyMs - info2.AvgLatencyMs,
		SharedCapabilities: shared,
		UniqueToFirst:      unique1,
		UniqueToSecond:     unique2,
		TokenLimitDiff:     info1.MaxTokens - info2.MaxTokens,
	}, nil
}

// RecommendModel suggests the best model for a given content type and constraints
type ModelConstraints struct {
	MaxCostPerThousandTokens float64           // Maximum cost per 1K tokens
	MaxLatencyMs             int               // Maximum acceptable latency
	RequiredCapabilities     []ModelCapability // Required capabilities
	MinTokenLimit            int               // Minimum token limit required
}

func RecommendModel(contentType ContentType, constraints ModelConstraints) []ModelType {
	var recommendations []ModelType

	// Content type specific capability requirements
	requiredCaps := constraints.RequiredCapabilities
	switch contentType {
	case TechnicalDoc:
		requiredCaps = append(requiredCaps, CodeAnalysis, StructuredOutput)
	case CreativeWriting:
		requiredCaps = append(requiredCaps, GeneralPurpose)
	case CodeSnippet:
		requiredCaps = append(requiredCaps, CodeAnalysis)
	case LegalDocument:
		requiredCaps = append(requiredCaps, HighAccuracy, SemanticAnalysis)
	case AcademicPaper:
		requiredCaps = append(requiredCaps, HighAccuracy, LongContext)
	case BusinessReport:
		requiredCaps = append(requiredCaps, StructuredOutput, SemanticAnalysis)
	case SocialMediaContent:
		requiredCaps = append(requiredCaps, FastResponse)
	}

	// Evaluate each model
	for modelType, info := range ModelRegistry {
		// Check cost constraint
		if info.Cost.InputPerThousandTokens > constraints.MaxCostPerThousandTokens {
			continue
		}

		// Check latency constraint
		if info.AvgLatencyMs > constraints.MaxLatencyMs {
			continue
		}

		// Check token limit
		if info.MaxTokens < constraints.MinTokenLimit {
			continue
		}

		// Check required capabilities
		hasAllCaps := true
		for _, reqCap := range requiredCaps {
			found := false
			for _, cap := range info.Capabilities {
				if cap == reqCap {
					found = true
					break
				}
			}
			if !found {
				hasAllCaps = false
				break
			}
		}

		if hasAllCaps {
			recommendations = append(recommendations, modelType)
		}
	}

	return recommendations
}

// GetModelInfo returns information about a specific model
func GetModelInfo(modelType ModelType) (ModelInfo, bool) {
	info, exists := ModelRegistry[modelType]
	return info, exists
}

// GetModelsByCapability returns all models that have a specific capability
func GetModelsByCapability(capability ModelCapability) []ModelInfo {
	var models []ModelInfo
	for _, model := range ModelRegistry {
		for _, cap := range model.Capabilities {
			if cap == capability {
				models = append(models, model)
				break
			}
		}
	}
	return models
}

// GetModelsByProvider returns all models for a specific provider
func GetModelsByProvider(provider Provider) []ModelInfo {
	var models []ModelInfo
	for _, model := range ModelRegistry {
		if model.Provider == provider {
			models = append(models, model)
		}
	}
	return models
}

// NewModelConfig creates a ModelConfig with default parameters for the specified model
func NewModelConfig(modelType ModelType, apiKey string) ModelConfig {
	info, exists := ModelRegistry[modelType]
	if !exists {
		return ModelConfig{
			Model:  string(modelType),
			APIKey: apiKey,
		}
	}

	return ModelConfig{
		Model:      string(modelType),
		APIKey:     apiKey,
		Parameters: info.Parameters,
	}
}
