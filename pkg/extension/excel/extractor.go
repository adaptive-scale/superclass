package excel

import (
	"strings"

	"github.com/unidoc/unioffice/spreadsheet"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(path string) (string, error) {
	wb, err := spreadsheet.Open(path)
	if err != nil {
		return "", err
	}
	defer wb.Close()

	var result strings.Builder

	// Process each sheet
	for _, sheet := range wb.Sheets() {
		result.WriteString("Sheet: " + sheet.Name() + "\n")

		// Process each row
		for _, row := range sheet.Rows() {
			var rowTexts []string

			// Process each cell in the row
			for _, cell := range row.Cells() {
				text := cell.GetString()
				if text != "" {
					rowTexts = append(rowTexts, text)
				}
			}

			// Add row text if not empty
			if len(rowTexts) > 0 {
				result.WriteString(strings.Join(rowTexts, "\t"))
				result.WriteString("\n")
			}
		}

		result.WriteString("\n") // Add extra newline between sheets
	}

	return result.String(), nil
}

func (e *Extractor) SupportedExtensions() []string {
	return []string{".xlsx", ".xlsm"}
}
