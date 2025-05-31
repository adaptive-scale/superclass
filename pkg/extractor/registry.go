package extractor

import (
	"fmt"
	"strings"
	"sync"
)

// TextExtractor defines the interface that all extractors must implement
type TextExtractor interface {
	// Extract extracts text from a file at the given path
	Extract(path string) (string, error)
	// SupportedExtensions returns a list of file extensions this extractor supports
	SupportedExtensions() []string
}

// Registry manages the registered text extractors
type Registry struct {
	mu         sync.RWMutex
	extractors map[string]TextExtractor // map of extension to extractor
}

// NewRegistry creates a new Registry instance
func NewRegistry() *Registry {
	return &Registry{
		extractors: make(map[string]TextExtractor),
	}
}

// Register adds a new TextExtractor to the registry
func (r *Registry) Register(extractor TextExtractor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, ext := range extractor.SupportedExtensions() {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if _, exists := r.extractors[ext]; exists {
			return fmt.Errorf("extractor for extension %s is already registered", ext)
		}
		r.extractors[ext] = extractor
	}
	return nil
}

// Get returns the registered TextExtractor for the given file extension
func (r *Registry) Get(extension string) (TextExtractor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	extension = strings.ToLower(extension)
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	extractor, exists := r.extractors[extension]
	if !exists {
		return nil, fmt.Errorf("no extractor registered for extension: %s", extension)
	}
	return extractor, nil
}

// GetSupportedExtensions returns a list of all supported file extensions
func (r *Registry) GetSupportedExtensions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	extensions := make([]string, 0, len(r.extractors))
	for ext := range r.extractors {
		extensions = append(extensions, ext)
	}
	return extensions
}

// DefaultRegistry is the default global registry
var DefaultRegistry = NewRegistry()
