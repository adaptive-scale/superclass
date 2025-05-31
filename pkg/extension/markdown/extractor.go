package markdown

import (
	"io/ioutil"
	"regexp"
	"strings"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	text := string(content)

	// Remove code blocks
	text = regexp.MustCompile("```[\\s\\S]*?```").ReplaceAllString(text, "")
	text = regexp.MustCompile("`[^`]*`").ReplaceAllString(text, "")

	// Remove links but keep link text
	text = regexp.MustCompile("\\[([^\\]]+)\\]\\([^)]+\\)").ReplaceAllString(text, "$1")

	// Remove images
	text = regexp.MustCompile("!\\[[^\\]]*\\]\\([^)]+\\)").ReplaceAllString(text, "")

	// Remove headers
	text = regexp.MustCompile("^#{1,6}\\s+(.+)$").ReplaceAllString(text, "$1")

	// Remove emphasis markers
	text = regexp.MustCompile("[*_]{1,3}([^*_]+)[*_]{1,3}").ReplaceAllString(text, "$1")

	// Remove HTML tags
	text = regexp.MustCompile("<[^>]+>").ReplaceAllString(text, "")

	// Remove horizontal rules
	text = regexp.MustCompile("^[-*_]{3,}\\s*$").ReplaceAllString(text, "")

	// Clean up whitespace
	text = regexp.MustCompile("\\s+").ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text, nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".md", ".markdown"}
}
