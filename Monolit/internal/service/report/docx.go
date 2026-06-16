package report

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

func generateDOCXReport(data ReportData) ([]byte, error) {
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	files := map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         relsXML,
		"word/document.xml":   documentXML(data),
	}

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return nil, fmt.Errorf("create docx part: %w", err)
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return nil, fmt.Errorf("write docx part: %w", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("close docx: %w", err)
	}

	return buffer.Bytes(), nil
}

func documentXML(data ReportData) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	docxParagraph(&b, "Отчет по звонку: "+data.Call.Title)
	docxParagraph(&b, "ID звонка: "+data.Call.ID.String())
	docxParagraph(&b, "Статус звонка: "+string(data.Call.Status))
	docxParagraph(&b, fmt.Sprintf("Длительность: %d сек.", data.Call.DurationSeconds))
	docxParagraph(&b, "Создан: "+data.Call.CreatedAt.Format(timeLayout))
	docxParagraph(&b, "Отчет создан: "+data.GeneratedAt.Format(timeLayout))
	docxParagraph(&b, "")
	docxParagraph(&b, "Анализ")
	docxParagraph(&b, "ID анализа: "+data.Analysis.ID.String())
	docxParagraph(&b, "Статус анализа: "+string(data.Analysis.Status))
	docxParagraph(&b, "Провайдер: "+data.Analysis.Provider)
	if data.Analysis.Model != nil {
		docxParagraph(&b, "Модель: "+*data.Analysis.Model)
	}

	for _, section := range data.Sections() {
		docxParagraph(&b, "")
		docxParagraph(&b, section.Title)
		for _, row := range section.Rows {
			if row.Label != "" && row.Value != "" {
				docxParagraph(&b, row.Label+": "+row.Value)
			} else if row.Value != "" {
				for _, paragraph := range splitParagraphs(row.Value) {
					docxParagraph(&b, paragraph)
				}
			} else if row.Label != "" {
				docxParagraph(&b, row.Label+":")
			}
			for _, item := range row.List {
				docxParagraph(&b, "• "+item)
			}
		}
	}

	b.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func docxParagraph(b *strings.Builder, text string) {
	b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
	b.WriteString(xmlEscape(text))
	b.WriteString(`</w:t></w:r></w:p>`)
}

func xmlEscape(value string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}

func splitParagraphs(value string) []string {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, line)
	}
	return out
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const relsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
