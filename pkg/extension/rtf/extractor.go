package rtf

import (
	"bytes"
	"io/ioutil"
	"strings"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	// Read the RTF file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Simple RTF text extraction
	// Remove RTF control sequences and extract plain text
	text := string(content)

	// Remove RTF header
	if idx := strings.Index(text, "\\rtf1"); idx >= 0 {
		text = text[idx:]
	}

	var buffer bytes.Buffer
	inControl := false
	inPlainText := false

	for i := 0; i < len(text); i++ {
		ch := text[i]

		switch ch {
		case '\\':
			inControl = true
			continue
		case '{', '}':
			continue
		case ' ':
			if inControl {
				inControl = false
			} else if inPlainText {
				buffer.WriteByte(ch)
			}
		default:
			if !inControl {
				if ch >= 32 && ch < 127 {
					buffer.WriteByte(ch)
					inPlainText = true
				} else if ch == '\n' || ch == '\r' {
					buffer.WriteByte(ch)
				}
			} else if ch == '\'' {
				// Skip hex-encoded character
				i += 2
				inControl = false
			} else if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				// Skip control word
				continue
			} else {
				inControl = false
			}
		}
	}

	return buffer.String(), nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".rtf"}
}
