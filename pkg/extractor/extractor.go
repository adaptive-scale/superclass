package extractor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/adaptive-scale/superclass/pkg/classifier"
	"github.com/adaptive-scale/superclass/pkg/extension/docx"
	"github.com/adaptive-scale/superclass/pkg/extension/epub"
	"github.com/adaptive-scale/superclass/pkg/extension/excel"
	"github.com/adaptive-scale/superclass/pkg/extension/html"
	"github.com/adaptive-scale/superclass/pkg/extension/image"
	"github.com/adaptive-scale/superclass/pkg/extension/markdown"
	"github.com/adaptive-scale/superclass/pkg/extension/odt"
	"github.com/adaptive-scale/superclass/pkg/extension/pdf"
	"github.com/adaptive-scale/superclass/pkg/extension/pptx"
	"github.com/adaptive-scale/superclass/pkg/extension/rtf"
	"github.com/adaptive-scale/superclass/pkg/extension/svg"
	log "github.com/sirupsen/logrus"
)

// ExtractResult contains both the extracted text and its classification
type ExtractResult struct {
	Text           string
	Classification *classifier.Classification
}

func init() {
	log.Debug("Initializing default registry with built-in extractors")
	// Register all built-in extractors
	DefaultRegistry.Register(pdf.NewExtractor())
	DefaultRegistry.Register(image.NewExtractor())
	DefaultRegistry.Register(docx.NewExtractor())
	DefaultRegistry.Register(pptx.NewExtractor())
	DefaultRegistry.Register(rtf.NewExtractor())
	DefaultRegistry.Register(odt.NewExtractor())
	DefaultRegistry.Register(html.NewExtractor())
	DefaultRegistry.Register(markdown.NewExtractor())
	DefaultRegistry.Register(epub.NewExtractor())
	DefaultRegistry.Register(excel.NewExtractor())
	DefaultRegistry.Register(svg.NewExtractor())
	log.Debug("All built-in extractors registered successfully")
}

// ExtractText extracts text from a file using the appropriate registered extractor
func ExtractText(path string) (string, error) {
	logger := log.WithFields(log.Fields{
		"function": "ExtractText",
		"path":     path,
	})
	logger.Debug("Starting text extraction")

	ext := strings.ToLower(filepath.Ext(path))
	logger.WithField("extension", ext).Debug("Detected file extension")

	// Special case for plain text files
	if ext == ".txt" {
		logger.Debug("Processing plain text file")
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			logger.WithError(err).Error("Failed to read text file")
			return "", err
		}
		logger.WithField("bytes_read", len(bytes)).Debug("Text file read successfully")
		return string(bytes), err
	}

	// Get the appropriate extractor from the registry
	logger.Debug("Looking up extractor from registry")
	extractor, err := DefaultRegistry.Get(ext)
	if err != nil {
		logger.WithError(err).Error("Failed to get extractor")
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}

	logger.Debug("Starting extraction with appropriate extractor")
	text, err := extractor.Extract(path)
	if err != nil {
		logger.WithError(err).Error("Extraction failed")
		return "", err
	}

	logger.WithFields(log.Fields{
		"chars_extracted": len(text),
		"lines_extracted": len(strings.Split(text, "\n")),
	}).Debug("Text extraction completed successfully")
	return text, nil
}

// ExtractAndClassify extracts text from a file and classifies it using the specified model
func ExtractAndClassify(path string, provider classifier.Provider, config classifier.ModelConfig) (*ExtractResult, error) {
	return ExtractAndClassifyWithOptions(path, provider, config, classifier.ClassificationOptions{})
}

// ExtractAndClassifyWithOptions extracts text from a file and classifies it using the specified model and options
func ExtractAndClassifyWithOptions(path string, provider classifier.Provider, config classifier.ModelConfig, options classifier.ClassificationOptions) (*ExtractResult, error) {
	logger := log.WithFields(log.Fields{
		"function":       "ExtractAndClassifyWithOptions",
		"path":           path,
		"provider":       provider,
		"model":          config.Model,
		"has_categories": len(options.Categories) > 0,
	})
	logger.Debug("Starting extraction and classification")

	// First extract the text
	logger.Debug("Extracting text from file")
	text, err := ExtractText(path)
	if err != nil {
		logger.WithError(err).Error("Text extraction failed")
		return nil, fmt.Errorf("text extraction failed: %w", err)
	}
	logger.WithField("text_length", len(text)).Debug("Text extraction completed")

	// Create classifier for the specified provider
	logger.Debug("Creating classifier instance")
	clf, err := classifier.NewClassifier(provider, config)
	if err != nil {
		logger.WithError(err).Error("Failed to create classifier")
		return nil, fmt.Errorf("failed to create classifier: %w", err)
	}

	// Classify the content with options
	logger.Debug("Starting content classification")
	classification, err := clf.ClassifyWithOptions(text, options)
	if err != nil {
		logger.WithError(err).Error("Classification failed")
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	logger.WithFields(log.Fields{
		"category":        classification.Category,
		"confidence":      classification.Confidence,
		"keywords_count":  len(classification.Keywords),
		"used_categories": options.Categories,
	}).Debug("Classification completed successfully")

	return &ExtractResult{
		Text:           text,
		Classification: classification,
	}, nil
}

// GetSupportedFormats returns a list of all supported file formats
func GetSupportedFormats() []string {
	log.Debug("Retrieving list of supported formats")
	formats := DefaultRegistry.GetSupportedExtensions()
	log.WithField("formats_count", len(formats)).Debug("Retrieved supported formats")
	return formats
}
