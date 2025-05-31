package pptx

import (
	"bytes"

	"github.com/unidoc/unioffice/presentation"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	ppt, err := presentation.Open(path)
	if err != nil {
		return "", err
	}
	defer ppt.Close()

	var buffer bytes.Buffer
	for _, slide := range ppt.Slides() {
		// Extract text from text boxes
		for _, textBox := range slide.GetTextBoxes() {
			for _, para := range textBox.X().TxBody.P {
				for _, run := range para.EG_TextRun {
					if run.R != nil && run.R.T != "" {
						buffer.WriteString(run.R.T)
						buffer.WriteString(" ")
					}
				}
			}
			buffer.WriteString("\n")
		}

		// Extract text from placeholders
		for _, ph := range slide.PlaceHolders() {
			for _, para := range ph.Paragraphs() {
				for _, run := range para.X().EG_TextRun {
					if run.R != nil && run.R.T != "" {
						buffer.WriteString(run.R.T)
						buffer.WriteString(" ")
					}
				}
			}
			buffer.WriteString("\n")
		}
	}

	return buffer.String(), nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".pptx"}
}
