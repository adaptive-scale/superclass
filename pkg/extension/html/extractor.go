package html

import (
	"bytes"
	"io/ioutil"
	"strings"

	"golang.org/x/net/html"
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

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	var result strings.Builder
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				result.WriteString(text)
				result.WriteString(" ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}

	extractText(doc)
	return strings.TrimSpace(result.String()), nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".html", ".htm"}
}
