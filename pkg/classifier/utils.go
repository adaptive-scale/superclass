package classifier

import (
	"strings"
)

// ProviderFromString converts a string to a Provider type
func ProviderFromString(provider string) Provider {
	switch strings.ToLower(provider) {
	case "openai":
		return OpenAI
	case "anthropic":
		return Anthropic
	case "azure":
		return Azure
	default:
		return OpenAI
	}
}

// CompareClassifications compares two classifications and returns similarity metrics
func CompareClassifications(a, b *Classification) *ComparisonResult {
	result := &ComparisonResult{
		CategoryMatch:     a.Category == b.Category,
		ConfidenceDiff:    a.Confidence - b.Confidence,
		KeywordOverlap:    calculateKeywordOverlap(a.Keywords, b.Keywords),
		SummarySimilarity: calculateSimilarity(a.Summary, b.Summary),
	}
	return result
}

// ComparisonResult contains metrics comparing two classifications
type ComparisonResult struct {
	CategoryMatch     bool    `json:"category_match"`
	ConfidenceDiff    float64 `json:"confidence_diff"`
	KeywordOverlap    float64 `json:"keyword_overlap"`
	SummarySimilarity float64 `json:"summary_similarity"`
}

// calculateKeywordOverlap calculates the Jaccard similarity between two keyword sets
func calculateKeywordOverlap(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Create sets
	setA := make(map[string]bool)
	setB := make(map[string]bool)
	for _, k := range a {
		setA[strings.ToLower(k)] = true
	}
	for _, k := range b {
		setB[strings.ToLower(k)] = true
	}

	// Calculate intersection and union
	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection

	return float64(intersection) / float64(union)
}

// calculateSimilarity calculates a simple similarity score between two strings
func calculateSimilarity(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	// Split into words
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)

	// Create word frequency maps
	freqA := make(map[string]int)
	freqB := make(map[string]int)
	for _, word := range wordsA {
		freqA[word]++
	}
	for _, word := range wordsB {
		freqB[word]++
	}

	// Calculate cosine similarity
	dotProduct := 0
	for word, countA := range freqA {
		if countB, exists := freqB[word]; exists {
			dotProduct += countA * countB
		}
	}

	// Calculate magnitudes
	magA := 0
	for _, count := range freqA {
		magA += count * count
	}
	magB := 0
	for _, count := range freqB {
		magB += count * count
	}

	if magA == 0 || magB == 0 {
		return 0.0
	}

	return float64(dotProduct) / (float64(magA) * float64(magB))
}
