package svg

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SVGElement represents an SVG element that can contain text
type SVGElement struct {
	XMLName      xml.Name
	Content      string           `xml:",chardata"`
	Title        string           `xml:"title"`
	Desc         string           `xml:"desc"`
	Tspan        []string         `xml:"tspan"`
	G            []SVGElement     `xml:"g"`
	TextElements []SVGTextElement `xml:"text"`
}

// SVGTextElement represents a text element in SVG
type SVGTextElement struct {
	Content string   `xml:",chardata"`
	Tspan   []string `xml:"tspan"`
}

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	logger := log.WithFields(log.Fields{
		"function": "Extract",
		"path":     path,
	})
	logger.Debug("Starting SVG text extraction")

	// Read the SVG file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logger.WithError(err).Error("Failed to read SVG file")
		return "", fmt.Errorf("failed to read SVG file: %w", err)
	}

	// Parse the SVG XML
	var svg SVGElement
	if err := xml.Unmarshal(content, &svg); err != nil {
		logger.WithError(err).Error("Failed to parse SVG XML")
		return "", fmt.Errorf("failed to parse SVG XML: %w", err)
	}

	// Extract text from all elements
	var textBuilder strings.Builder

	// Add title if present
	if svg.Title != "" {
		textBuilder.WriteString("Title: ")
		textBuilder.WriteString(svg.Title)
		textBuilder.WriteString("\n\n")
	}

	// Add description if present
	if svg.Desc != "" {
		textBuilder.WriteString("Description: ")
		textBuilder.WriteString(svg.Desc)
		textBuilder.WriteString("\n\n")
	}

	// Extract text from all elements recursively
	extractText(&svg, &textBuilder)

	result := textBuilder.String()
	logger.WithField("extracted_length", len(result)).Debug("SVG text extraction completed")

	return result, nil
}

func extractText(element *SVGElement, builder *strings.Builder) {
	// Extract direct text content
	if text := strings.TrimSpace(element.Content); text != "" {
		builder.WriteString(text)
		builder.WriteString("\n")
	}

	// Extract text from tspan elements
	for _, span := range element.Tspan {
		if text := strings.TrimSpace(span); text != "" {
			builder.WriteString(text)
			builder.WriteString("\n")
		}
	}

	// Extract text from text elements
	for _, textElement := range element.TextElements {
		if text := strings.TrimSpace(textElement.Content); text != "" {
			builder.WriteString(text)
			builder.WriteString("\n")
		}
		// Extract from nested tspans
		for _, span := range textElement.Tspan {
			if text := strings.TrimSpace(span); text != "" {
				builder.WriteString(text)
				builder.WriteString("\n")
			}
		}
	}

	// Recursively process group elements
	for _, group := range element.G {
		extractText(&group, builder)
	}
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".svg"}
}
