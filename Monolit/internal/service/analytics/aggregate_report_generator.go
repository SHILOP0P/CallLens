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
	Summary                string                                `json:"summary"`
	KeyFindings            []string                              `json:"key_findings"`
	RecurringIssues        []string                              `json:"recurring_issues"`
	Strengths              []string                              `json:"strengths"`
	Risks                  []string                              `json:"risks"`
	PriorityActions        []string                              `json:"priority_actions"`
	ManagerRecommendations []string                              `json:"manager_recommendations"`
	Confidence             any                                   `json:"confidence"`
	SourceSummary          models.AggregateAnalysisSourceSummary `json:"source_summary"`
	AggregateStatistics    aggregateReportStatistics             `json:"aggregate_statistics"`
	CoverageNote           string                                `json:"coverage_note"`
}

// aggregateReportStatistics mirrors the deterministic dataset saved with a deep
// analysis. Keeping it in the export means that every report format exposes
// statistics calculated from the complete source set, not only the AI's
// representative-call narrative.
type aggregateReportStatistics struct {
	ScoreSummary       models.AggregateAnalysisScoreSummary      `json:"score_summary"`
	IssueCoverage      []models.AggregateAnalysisFrequency       `json:"issue_coverage"`
	WeakCriteria       []models.AggregateAnalysisCriterionMetric `json:"weak_criteria"`
	BusinessOutcomes   []models.AggregateAnalysisFrequency       `json:"business_outcomes"`
	LostReasons        []models.AggregateAnalysisFrequency       `json:"lost_reasons"`
	CustomerObjections []models.AggregateAnalysisFrequency       `json:"customer_objections"`
	Risks              []models.AggregateAnalysisFrequency       `json:"risks"`
	Topics             []models.AggregateAnalysisFrequency       `json:"topics"`
	NextStepSummary    models.AggregateAnalysisNextStepSummary   `json:"next_step_summary"`
	AttentionCalls     []models.AggregateAnalysisCallEvidence    `json:"attention_calls"`
	StrongCalls        []models.AggregateAnalysisCallEvidence    `json:"strong_calls"`
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
	writeMarkdownSection(b, "Data coverage", aggregateCoverageLines(a, result))
	writeMarkdownSection(b, "Score summary", aggregateScoreLines(result.AggregateStatistics.ScoreSummary))
	writeMarkdownSection(b, "Issue coverage", aggregateFrequencyLines(result.AggregateStatistics.IssueCoverage))
	writeMarkdownSection(b, "Weak criteria", aggregateWeakCriteriaLines(result.AggregateStatistics.WeakCriteria))
	writeMarkdownSection(b, "Business outcomes", aggregateFrequencyLines(result.AggregateStatistics.BusinessOutcomes))
	writeMarkdownSection(b, "Lost reasons", aggregateFrequencyLines(result.AggregateStatistics.LostReasons))
	writeMarkdownSection(b, "Customer objections", aggregateFrequencyLines(result.AggregateStatistics.CustomerObjections))
	writeMarkdownSection(b, "Risk metrics", aggregateFrequencyLines(result.AggregateStatistics.Risks))
	writeMarkdownSection(b, "Topics", aggregateFrequencyLines(result.AggregateStatistics.Topics))
	writeMarkdownSection(b, "Next-step quality", aggregateNextStepLines(result.AggregateStatistics.NextStepSummary))
	writeMarkdownSection(b, "Attention calls", aggregateCallEvidenceLines(result.AggregateStatistics.AttentionCalls))
	writeMarkdownSection(b, "Strong calls", aggregateCallEvidenceLines(result.AggregateStatistics.StrongCalls))
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
		{"Coverage note", result.CoverageNote},
		{"Analyzed calls", result.SourceSummary.AnalyzedCalls},
		{"Included in statistics", result.SourceSummary.IncludedInStatistics},
		{"All analyzed calls used", result.SourceSummary.AllAnalyzedCallsUsed},
		{"Representative calls", result.SourceSummary.RepresentativeCalls},
		{"Source set hash", result.SourceSummary.SourceSetHash},
		{"Calls with score", result.AggregateStatistics.ScoreSummary.CallsWithScore},
		{"Average score", formatOptionalFloat(result.AggregateStatistics.ScoreSummary.Average)},
		{"Minimum score", formatOptionalFloat(result.AggregateStatistics.ScoreSummary.Min)},
		{"Maximum score", formatOptionalFloat(result.AggregateStatistics.ScoreSummary.Max)},
		{"Low-score calls", result.AggregateStatistics.ScoreSummary.LowCount},
		{"Medium-score calls", result.AggregateStatistics.ScoreSummary.MediumCount},
		{"High-score calls", result.AggregateStatistics.ScoreSummary.HighCount},
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
	for _, sheet := range []struct {
		name string
		rows [][]any
	}{
		{"Issue coverage", aggregateFrequencyRows(result.AggregateStatistics.IssueCoverage)},
		{"Weak criteria", aggregateWeakCriteriaRows(result.AggregateStatistics.WeakCriteria)},
		{"Business outcomes", aggregateFrequencyRows(result.AggregateStatistics.BusinessOutcomes)},
		{"Lost reasons", aggregateFrequencyRows(result.AggregateStatistics.LostReasons)},
		{"Objections", aggregateFrequencyRows(result.AggregateStatistics.CustomerObjections)},
		{"Risk metrics", aggregateFrequencyRows(result.AggregateStatistics.Risks)},
		{"Topics", aggregateFrequencyRows(result.AggregateStatistics.Topics)},
		{"Next steps", aggregateNextStepRows(result.AggregateStatistics.NextStepSummary)},
		{"Attention calls", aggregateCallEvidenceRows(result.AggregateStatistics.AttentionCalls)},
		{"Strong calls", aggregateCallEvidenceRows(result.AggregateStatistics.StrongCalls)},
	} {
		if err := addAggregateSheet(file, sheet.name, sheet.rows); err != nil {
			return nil, err
		}
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

func aggregateCoverageLines(analysis models.AggregateAnalysis, result aggregateResult) []string {
	summary := result.SourceSummary
	if summary.AnalyzedCalls == 0 && summary.IncludedInStatistics == 0 && summary.SourceSetHash == "" && result.CoverageNote == "" && analysis.SourceCallsCount == 0 {
		return nil
	}
	analyzedCalls := summary.AnalyzedCalls
	if analyzedCalls == 0 {
		analyzedCalls = analysis.SourceCallsCount
	}
	lines := []string{
		fmt.Sprintf("Analyzed calls: %d", analyzedCalls),
		fmt.Sprintf("Included in statistics: %d", summary.IncludedInStatistics),
		fmt.Sprintf("All analyzed calls used: %t", summary.AllAnalyzedCallsUsed),
		fmt.Sprintf("Representative calls used by AI: %d", summary.RepresentativeCalls),
	}
	if summary.SourceSetHash != "" {
		lines = append(lines, "Source set hash: "+summary.SourceSetHash)
	}
	if result.CoverageNote != "" {
		lines = append(lines, result.CoverageNote)
	}
	return lines
}

func aggregateScoreLines(summary models.AggregateAnalysisScoreSummary) []string {
	if summary.CallsWithScore == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("Calls with score: %d", summary.CallsWithScore),
		"Average: " + formatOptionalFloat(summary.Average),
		"Minimum: " + formatOptionalFloat(summary.Min),
		"Maximum: " + formatOptionalFloat(summary.Max),
		fmt.Sprintf("Score distribution — low: %d, medium: %d, high: %d", summary.LowCount, summary.MediumCount, summary.HighCount),
	}
}

func aggregateFrequencyLines(items []models.AggregateAnalysisFrequency) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		line := fmt.Sprintf("%s — %d calls (%.1f%%)", fallback(item.Title, item.Code), item.Count, item.Share*100)
		if len(item.SampleCallUUIDs) > 0 {
			line += "; sample calls: " + strings.Join(item.SampleCallUUIDs, ", ")
		}
		lines = append(lines, line)
	}
	return lines
}

func aggregateWeakCriteriaLines(items []models.AggregateAnalysisCriterionMetric) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		line := fmt.Sprintf("%s — weak in %d of %d applicable calls (%.1f%%); missed: %d, partial: %d, unclear: %d", fallback(item.Title, item.Code), item.WeakCalls, item.ApplicableCalls, item.WeakShare*100, item.MissedCalls, item.PartiallyMetCalls, item.UnclearCalls)
		if item.AveragePointsShare != nil {
			line += fmt.Sprintf("; average points: %.1f%%", *item.AveragePointsShare*100)
		}
		if len(item.SampleCallUUIDs) > 0 {
			line += "; sample calls: " + strings.Join(item.SampleCallUUIDs, ", ")
		}
		lines = append(lines, line)
	}
	return lines
}

func aggregateNextStepLines(summary models.AggregateAnalysisNextStepSummary) []string {
	if summary.CallsWithNextStep == 0 && summary.CallsMissingNextStep == 0 && summary.CallsWithSpecificNextStep == 0 && summary.CallsMissingSpecificStep == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("Calls with next step: %d; missing next step: %d (%.1f%%)", summary.CallsWithNextStep, summary.CallsMissingNextStep, summary.MissingNextStepShare*100),
		fmt.Sprintf("Calls with specific next step: %d; missing specific step: %d (%.1f%%)", summary.CallsWithSpecificNextStep, summary.CallsMissingSpecificStep, summary.MissingSpecificStepShare*100),
	}
}

func aggregateCallEvidenceLines(items []models.AggregateAnalysisCallEvidence) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		line := fmt.Sprintf("%s (%s)", fallback(item.Title, item.CallUUID.String()), item.CallUUID)
		if item.Score != nil {
			line += fmt.Sprintf(" — score %.2f", *item.Score)
		}
		if item.Summary != "" {
			line += "; " + item.Summary
		}
		if len(item.IssueCodes) > 0 {
			line += "; issues: " + strings.Join(item.IssueCodes, ", ")
		}
		lines = append(lines, line)
	}
	return lines
}

func formatOptionalFloat(value *float64) string {
	if value == nil {
		return "No data"
	}
	return fmt.Sprintf("%.2f", *value)
}

func aggregateFrequencyRows(items []models.AggregateAnalysisFrequency) [][]any {
	rows := [][]any{{"Code", "Title", "Calls", "Share", "Sample call UUIDs"}}
	for _, item := range items {
		rows = append(rows, []any{item.Code, item.Title, item.Count, item.Share, strings.Join(item.SampleCallUUIDs, ", ")})
	}
	return rows
}

func aggregateWeakCriteriaRows(items []models.AggregateAnalysisCriterionMetric) [][]any {
	rows := [][]any{{"Code", "Title", "Applicable", "Weak", "Weak share", "Avg. points share", "Missed", "Partial", "Unclear", "Sample call UUIDs"}}
	for _, item := range items {
		rows = append(rows, []any{item.Code, item.Title, item.ApplicableCalls, item.WeakCalls, item.WeakShare, formatOptionalFloat(item.AveragePointsShare), item.MissedCalls, item.PartiallyMetCalls, item.UnclearCalls, strings.Join(item.SampleCallUUIDs, ", ")})
	}
	return rows
}

func aggregateNextStepRows(summary models.AggregateAnalysisNextStepSummary) [][]any {
	return [][]any{
		{"Metric", "Value"},
		{"Calls with next step", summary.CallsWithNextStep},
		{"Calls with specific next step", summary.CallsWithSpecificNextStep},
		{"Calls missing next step", summary.CallsMissingNextStep},
		{"Calls missing specific step", summary.CallsMissingSpecificStep},
		{"Missing next step share", summary.MissingNextStepShare},
		{"Missing specific step share", summary.MissingSpecificStepShare},
	}
}

func aggregateCallEvidenceRows(items []models.AggregateAnalysisCallEvidence) [][]any {
	rows := [][]any{{"Call UUID", "Created at", "Title", "Score", "Summary", "Issue codes"}}
	for _, item := range items {
		rows = append(rows, []any{item.CallUUID.String(), item.CreatedAt.Format(timeLayout), item.Title, formatOptionalFloat(item.Score), item.Summary, strings.Join(item.IssueCodes, ", ")})
	}
	return rows
}

func addAggregateSheet(file *excelize.File, name string, rows [][]any) error {
	if _, err := file.NewSheet(name); err != nil {
		return err
	}
	setSheetRows(file, name, rows)
	return file.SetColWidth(name, "C", "J", 18)
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
