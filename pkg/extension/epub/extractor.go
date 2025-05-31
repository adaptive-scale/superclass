package epub

import (
	"archive/zip"
	"encoding/xml"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

type container struct {
	XMLName  xml.Name `xml:"container"`
	Rootfile struct {
		Path string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type opf struct {
	XMLName xml.Name `xml:"package"`
	Spine   struct {
		Items []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
	Manifest struct {
		Items []struct {
			ID   string `xml:"id,attr"`
			Href string `xml:"href,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
}

func (e *Extractor) Extract(path string) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// First, read container.xml to find the OPF file
	var containerFile *zip.File
	for _, file := range reader.File {
		if file.Name == "META-INF/container.xml" {
			containerFile = file
			break
		}
	}

	if containerFile == nil {
		return "", err
	}

	rc, err := containerFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var container container
	if err := xml.NewDecoder(rc).Decode(&container); err != nil {
		return "", err
	}

	// Read the OPF file
	var opfFile *zip.File
	for _, file := range reader.File {
		if file.Name == container.Rootfile.Path {
			opfFile = file
			break
		}
	}

	if opfFile == nil {
		return "", err
	}

	rc, err = opfFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var opf opf
	if err := xml.NewDecoder(rc).Decode(&opf); err != nil {
		return "", err
	}

	// Extract text from each content file
	var result strings.Builder
	opfDir := filepath.Dir(container.Rootfile.Path)

	for _, spineItem := range opf.Spine.Items {
		var href string
		for _, manifestItem := range opf.Manifest.Items {
			if manifestItem.ID == spineItem.IDRef {
				href = manifestItem.Href
				break
			}
		}

		if href == "" {
			continue
		}

		contentPath := filepath.Join(opfDir, href)
		var contentFile *zip.File
		for _, file := range reader.File {
			if file.Name == contentPath {
				contentFile = file
				break
			}
		}

		if contentFile == nil {
			continue
		}

		rc, err = contentFile.Open()
		if err != nil {
			continue
		}

		doc, err := html.Parse(rc)
		rc.Close()
		if err != nil {
			continue
		}

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
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String()), nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".epub"}
}
