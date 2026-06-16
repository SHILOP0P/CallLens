package report

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

func generateXLSXReport(data ReportData) ([]byte, error) {
	file := excelize.NewFile()
	defer file.Close()
	analysis := data.StructuredAnalysis()

	metaSheet := "Метаданные"
	file.SetSheetName("Sheet1", metaSheet)
	setRows(file, metaSheet, [][]any{
		{"Поле", "Значение"},
		{"ID звонка", data.Call.ID.String()},
		{"Название", data.Call.Title},
		{"Статус звонка", string(data.Call.Status)},
		{"Длительность, сек.", data.Call.DurationSeconds},
		{"Создан", data.Call.CreatedAt.Format(timeLayout)},
		{"Отчет создан", data.GeneratedAt.Format(timeLayout)},
		{"ID анализа", data.Analysis.ID.String()},
		{"Статус анализа", string(data.Analysis.Status)},
		{"Провайдер", data.Analysis.Provider},
		{"Модель", optionalString(data.Analysis.Model)},
	})

	analysisSheet := "Анализ"
	if _, err := file.NewSheet(analysisSheet); err != nil {
		return nil, fmt.Errorf("create analysis sheet: %w", err)
	}
	setRows(file, analysisSheet, sectionRows(data.Sections()))

	if len(analysis.ClientQuestions) > 0 {
		questionsSheet := "Вопросы"
		if _, err := file.NewSheet(questionsSheet); err != nil {
			return nil, fmt.Errorf("create questions sheet: %w", err)
		}
		rows := [][]any{{"Вопрос", "Ответ менеджера", "Статус", "Цитаты"}}
		for _, question := range analysis.ClientQuestions {
			rows = append(rows, []any{
				question.Question,
				question.ManagerAnswer,
				answerStatusLabel(question.AnswerStatus),
				strings.Join(question.EvidenceQuotes, "\n"),
			})
		}
		setRows(file, questionsSheet, rows)
	}

	if len(analysis.CriteriaResults) > 0 {
		criteriaSheet := "Критерии"
		if _, err := file.NewSheet(criteriaSheet); err != nil {
			return nil, fmt.Errorf("create criteria sheet: %w", err)
		}
		rows := [][]any{{"Критерий", "Результат", "Цитаты"}}
		for _, criterion := range analysis.CriteriaResults {
			rows = append(rows, []any{
				criterion.InstructionTitle,
				criterion.Result,
				strings.Join(criterion.EvidenceQuotes, "\n"),
			})
		}
		setRows(file, criteriaSheet, rows)
	}

	if data.TranscriptionText != "" {
		transcriptionSheet := "Транскрипция"
		if _, err := file.NewSheet(transcriptionSheet); err != nil {
			return nil, fmt.Errorf("create transcription sheet: %w", err)
		}
		rows := [][]any{{"Строка", "Текст"}}
		for index, paragraph := range splitParagraphs(data.TranscriptionText) {
			rows = append(rows, []any{index + 1, paragraph})
		}
		setRows(file, transcriptionSheet, rows)
	}

	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, fmt.Errorf("generate xlsx report: %w", err)
	}

	return buffer.Bytes(), nil
}

func setRows(file *excelize.File, sheet string, rows [][]any) {
	for rowIndex, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+1)
		_ = file.SetSheetRow(sheet, cell, &row)
	}
	_ = file.SetColWidth(sheet, "A", "A", 24)
	_ = file.SetColWidth(sheet, "B", "B", 100)
	_ = file.SetColWidth(sheet, "C", "D", 60)
}

func sectionRows(sections []reportSection) [][]any {
	rows := [][]any{{"Раздел", "Поле", "Значение"}}
	for _, section := range sections {
		for _, row := range section.Rows {
			if row.Value != "" {
				rows = append(rows, []any{section.Title, row.Label, row.Value})
			}
			if len(row.List) > 0 {
				rows = append(rows, []any{section.Title, row.Label, strings.Join(row.List, "\n")})
			}
		}
	}
	return rows
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
