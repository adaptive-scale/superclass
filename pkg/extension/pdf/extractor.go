package pdf

import (
	"strings"

	"github.com/ledongthuc/pdf"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	f, r, err := pdf.Open(path)
	defer f.Close()
	if err != nil {
		return "", err
	}
	var textBuilder strings.Builder
	rNumPages := r.NumPage()
	for i := 1; i <= rNumPages; i++ {
		page := r.Page(i)
		content, _ := page.GetPlainText(nil)
		textBuilder.WriteString(content)
	}
	return textBuilder.String(), nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".pdf"}
}

// For backward compatibility
func ExtractTextFromPDF(path string) (string, error) {
	return NewExtractor().Extract(path)
}
