package docx

import (
	"github.com/unidoc/unioffice/document"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	doc, err := document.Open(path)
	if err != nil {
		return "", err
	}
	defer doc.Close()

	var text string
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			text += run.Text()
		}
		text += "\n"
	}
	return text, nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".docx"}
}
