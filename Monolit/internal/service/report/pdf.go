package report

import (
	"bytes"
	"fmt"
	"os"

	"github.com/jung-kurt/gofpdf"
)

func generatePDFReport(data ReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	fontPath, err := reportFontPath()
	if err != nil {
		return nil, err
	}
	pdf.AddUTF8Font("report", "", fontPath)
	pdf.SetFont("report", "", 12)
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	writePDFLine(pdf, 16, "Отчет по звонку: "+data.Call.Title)
	writePDFLine(pdf, 11, "ID звонка: "+data.Call.ID.String())
	writePDFLine(pdf, 11, "Статус звонка: "+string(data.Call.Status))
	writePDFLine(pdf, 11, fmt.Sprintf("Длительность: %d сек.", data.Call.DurationSeconds))
	writePDFLine(pdf, 11, "Создан: "+data.Call.CreatedAt.Format(timeLayout))
	writePDFLine(pdf, 11, "Отчет создан: "+data.GeneratedAt.Format(timeLayout))
	pdf.Ln(4)

	writePDFLine(pdf, 14, "Анализ")
	writePDFLine(pdf, 11, "ID анализа: "+data.Analysis.ID.String())
	writePDFLine(pdf, 11, "Статус анализа: "+string(data.Analysis.Status))
	writePDFLine(pdf, 11, "Провайдер: "+data.Analysis.Provider)
	if data.Analysis.Model != nil {
		writePDFLine(pdf, 11, "Модель: "+*data.Analysis.Model)
	}
	pdf.Ln(2)

	for _, section := range data.Sections() {
		pdf.Ln(4)
		writePDFLine(pdf, 14, section.Title)
		for _, row := range section.Rows {
			if row.Label != "" && row.Value != "" {
				writePDFBlock(pdf, row.Label+": "+row.Value)
			} else if row.Value != "" {
				writePDFBlock(pdf, row.Value)
			} else if row.Label != "" {
				writePDFBlock(pdf, row.Label+":")
			}
			for _, item := range row.List {
				writePDFBlock(pdf, "• "+item)
			}
			pdf.Ln(1)
		}
	}

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, fmt.Errorf("generate pdf report: %w", err)
	}

	return buffer.Bytes(), nil
}

func reportFontPath() (string, error) {
	candidates := []string{
		`C:\Windows\Fonts\arial.ttf`,
		`C:\Windows\Fonts\segoeui.ttf`,
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/dejavu/DejaVuSans.ttf",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("report pdf font not found")
}

func writePDFLine(pdf *gofpdf.Fpdf, size float64, text string) {
	pdf.SetFontSize(size)
	pdf.MultiCell(0, 7, text, "", "L", false)
}

func writePDFBlock(pdf *gofpdf.Fpdf, text string) {
	pdf.SetFontSize(10)
	for _, paragraph := range splitParagraphs(text) {
		pdf.MultiCell(0, 5, paragraph, "", "L", false)
	}
}
