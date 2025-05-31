package odt

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

type content struct {
	XMLName xml.Name `xml:"document-content"`
	Body    struct {
		Text struct {
			Ps []struct {
				Spans []struct {
					Text string `xml:",chardata"`
				} `xml:"span"`
				Text string `xml:",chardata"`
			} `xml:"p"`
		} `xml:"text"`
	} `xml:"body"`
}

func (e *Extractor) Extract(path string) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var contentXML *zip.File
	for _, file := range reader.File {
		if file.Name == "content.xml" {
			contentXML = file
			break
		}
	}

	if contentXML == nil {
		return "", err
	}

	rc, err := contentXML.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return "", err
	}

	var c content
	if err := xml.Unmarshal(buf.Bytes(), &c); err != nil {
		return "", err
	}

	var result strings.Builder
	for _, p := range c.Body.Text.Ps {
		result.WriteString(p.Text)
		for _, span := range p.Spans {
			result.WriteString(span.Text)
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".odt"}
}
