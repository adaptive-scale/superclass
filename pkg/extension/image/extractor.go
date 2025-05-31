package image

import (
	gosseract "github.com/otiai10/gosseract/v2"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	// Set the image path
	if err := client.SetImage(path); err != nil {
		return "", err
	}

	// Set additional OCR configurations for better accuracy
	client.SetLanguage("eng")                         // Use English language
	client.SetConfigFile("preserve_interword_spaces") // Preserve spacing between words

	// Perform OCR
	return client.Text()
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{
		".jpg", ".jpeg", ".png", ".gif",
		".bmp", ".tiff", ".tif", ".webp",
	}
}
