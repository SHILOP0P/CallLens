package analytics

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type AggregateReportData struct {
	Analysis    models.AggregateAnalysis
	GeneratedAt time.Time
}

type aggregateResult struct {
	Summary                string   `json:"summary"`
	KeyFindings            []string `json:"key_findings"`
	RecurringIssues        []string `json:"recurring_issues"`
	Strengths              []string `json:"strengths"`
	Risks                  []string `json:"risks"`
	PriorityActions        []string `json:"priority_actions"`
	ManagerRecommendations []string `json:"manager_recommendations"`
	Confidence             any      `json:"confidence"`
}

func generateAggregateReport(format models.ReportFormat, data AggregateReportData) ([]byte, error) {
	switch format {
	case models.ReportFormatMD:
		return generateAggregateMarkdownReport(data), nil
	case models.ReportFormatXLSX:
		return generateAggregateXLSXReport(data)
	case models.ReportFormatPDF:
		return generateAggregatePDFReport(data)
	case models.ReportFormatDOCX:
		return generateAggregateDOCXReport(data)
	default:
		return nil, models.ErrUnsupportedReportFormat
	}
}

func generateAggregateMarkdownReport(data AggregateReportData) []byte {
	result := parseAggregateResult(data.Analysis)
	var b bytes.Buffer
	writeAggregateMarkdown(&b, data, result, true)
	return b.Bytes()
}

func writeAggregateMarkdown(b *bytes.Buffer, data AggregateReportData, result aggregateResult, includeRaw bool) {
	a := data.Analysis
	fmt.Fprintln(b, "# Deep analysis report")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- Aggregate analysis ID: `%s`\n", a.ID)
	fmt.Fprintf(b, "- Scope: `%s`\n", a.Scope)
	fmt.Fprintf(b, "- Period: %s - %s\n", a.PeriodFrom.Format(timeLayout), a.PeriodTo.Format(timeLayout))
	fmt.Fprintf(b, "- Source calls count: %d\n", a.SourceCallsCount)
	fmt.Fprintf(b, "- Generated at: %s\n", data.GeneratedAt.Format(timeLayout))
	if a.Model != nil {
		fmt.Fprintf(b, "- Model: `%s`\n", *a.Model)
	}
	fmt.Fprintln(b)
	writeMarkdownSection(b, "Summary", []string{fallback(result.Summary, resultText(a))})
	writeMarkdownSection(b, "Key findings", result.KeyFindings)
	writeMarkdownSection(b, "Recurring issues", result.RecurringIssues)
	writeMarkdownSection(b, "Strengths", result.Strengths)
	writeMarkdownSection(b, "Risks", result.Risks)
	writeMarkdownSection(b, "Priority actions", result.PriorityActions)
	writeMarkdownSection(b, "Manager recommendations", result.ManagerRecommendations)
	if result.Confidence != nil {
		writeMarkdownSection(b, "Confidence", []string{fmt.Sprint(result.Confidence)})
	}
	if includeRaw && len(a.ResultJSON) > 0 {
		fmt.Fprintln(b, "## Raw JSON")
		fmt.Fprintln(b)
		fmt.Fprintln(b, "```json")
		_ = json.Indent(b, a.ResultJSON, "", "  ")
		fmt.Fprintln(b)
		fmt.Fprintln(b, "```")
	}
}

func writeMarkdownSection(b *bytes.Buffer, title string, items []string) {
	fmt.Fprintf(b, "## %s\n\n", title)
	if len(items) == 0 {
		fmt.Fprintln(b, "No data.")
		fmt.Fprintln(b)
		return
	}
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			fmt.Fprintf(b, "- %s\n", strings.TrimSpace(item))
		}
	}
	fmt.Fprintln(b)
}

func generateAggregateXLSXReport(data AggregateReportData) ([]byte, error) {
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	result := parseAggregateResult(data.Analysis)
	_ = file.SetSheetName("Sheet1", "Summary")
	setSheetRows(file, "Summary", [][]any{
		{"Field", "Value"},
		{"Aggregate analysis ID", data.Analysis.ID.String()},
		{"Scope", string(data.Analysis.Scope)},
		{"Period from", data.Analysis.PeriodFrom.Format(timeLayout)},
		{"Period to", data.Analysis.PeriodTo.Format(timeLayout)},
		{"Source calls count", data.Analysis.SourceCallsCount},
		{"Generated at", data.GeneratedAt.Format(timeLayout)},
		{"Summary", fallback(result.Summary, resultText(data.Analysis))},
		{"Confidence", fmt.Sprint(result.Confidence)},
	})
	for _, sheet := range []struct {
		name  string
		items []string
	}{
		{"Findings", result.KeyFindings},
		{"Issues", result.RecurringIssues},
		{"Actions", result.PriorityActions},
		{"Strengths", result.Strengths},
		{"Risks", result.Risks},
	} {
		if _, err := file.NewSheet(sheet.name); err != nil {
			return nil, err
		}
		rows := [][]any{{"Item"}}
		for _, item := range sheet.items {
			rows = append(rows, []any{item})
		}
		setSheetRows(file, sheet.name, rows)
	}
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func generateAggregateDOCXReport(data AggregateReportData) ([]byte, error) {
	result := parseAggregateResult(data.Analysis)
	var text bytes.Buffer
	writeAggregateMarkdown(&text, data, result, false)
	var doc strings.Builder
	doc.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	doc.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, line := range strings.Split(text.String(), "\n") {
		docxParagraph(&doc, strings.TrimPrefix(line, "- "))
	}
	doc.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>`)
	doc.WriteString(`</w:body></w:document>`)
	return zipDocx(doc.String())
}

func generateAggregatePDFReport(data AggregateReportData) ([]byte, error) {
	result := parseAggregateResult(data.Analysis)
	pdf := gofpdf.New("P", "mm", "A4", "")
	fontPath, err := reportFontPath()
	if err != nil {
		return nil, err
	}
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}
	pdf.AddUTF8FontFromBytes("report", "", fontBytes)
	pdf.SetFont("report", "", 12)
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()
	var text bytes.Buffer
	writeAggregateMarkdown(&text, data, result, false)
	for _, line := range strings.Split(text.String(), "\n") {
		pdf.MultiCell(0, 6, line, "", "L", false)
	}
	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func parseAggregateResult(analysis models.AggregateAnalysis) aggregateResult {
	var result aggregateResult
	if len(analysis.ResultJSON) > 0 && json.Unmarshal(analysis.ResultJSON, &result) == nil {
		return result
	}
	result.Summary = resultText(analysis)
	return result
}

func resultText(analysis models.AggregateAnalysis) string {
	if analysis.ResultText != nil && strings.TrimSpace(*analysis.ResultText) != "" {
		return strings.TrimSpace(*analysis.ResultText)
	}
	return "No structured result is available."
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func setSheetRows(file *excelize.File, sheet string, rows [][]any) {
	for rowIndex, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+1)
		_ = file.SetSheetRow(sheet, cell, &row)
	}
	_ = file.SetColWidth(sheet, "A", "A", 28)
	_ = file.SetColWidth(sheet, "B", "B", 100)
}

func aggregateReportFileName(analysis models.AggregateAnalysis, id uuid.UUID, format models.ReportFormat) string {
	return fmt.Sprintf("deep-analysis-%s-%s-%s-%s%s",
		analysis.Scope,
		analysis.PeriodFrom.Format("2006-01-02"),
		analysis.PeriodTo.Format("2006-01-02"),
		id.String(),
		reportFileExtension(format),
	)
}

func normalizeReportFormat(format models.ReportFormat) (models.ReportFormat, error) {
	switch format {
	case models.ReportFormatPDF, models.ReportFormatDOCX, models.ReportFormatMD, models.ReportFormatXLSX:
		return format, nil
	default:
		return "", models.ErrUnsupportedReportFormat
	}
}

func reportContentType(format models.ReportFormat) string {
	switch format {
	case models.ReportFormatPDF:
		return "application/pdf"
	case models.ReportFormatDOCX:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case models.ReportFormatMD:
		return "text/markdown; charset=utf-8"
	case models.ReportFormatXLSX:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

func reportFileExtension(format models.ReportFormat) string {
	switch format {
	case models.ReportFormatPDF:
		return ".pdf"
	case models.ReportFormatDOCX:
		return ".docx"
	case models.ReportFormatMD:
		return ".md"
	case models.ReportFormatXLSX:
		return ".xlsx"
	default:
		return ""
	}
}

func reportFontPath() (string, error) {
	for _, candidate := range []string{`C:\Windows\Fonts\arial.ttf`, `C:\Windows\Fonts\segoeui.ttf`, "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("report pdf font not found")
}

func zipDocx(document string) ([]byte, error) {
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`,
		"_rels/.rels":         `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`,
		"word/document.xml":   document,
	}
	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func docxParagraph(b *strings.Builder, text string) {
	b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
	_ = xml.EscapeText((*stringWriter)(b), []byte(text))
	b.WriteString(`</w:t></w:r></w:p>`)
}

type stringWriter strings.Builder

func (w *stringWriter) Write(p []byte) (int, error) {
	(*strings.Builder)(w).Write(p)
	return len(p), nil
}

const timeLayout = "2006-01-02 15:04:05 UTC"
