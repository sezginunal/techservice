package xlsx

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Write writes a single-sheet .xlsx workbook to w.
func Write(w io.Writer, sheet string, rows [][]string) error {
	name := sheetName(sheet)
	shared := []string{}
	index := map[string]int{}
	for _, row := range rows {
		for _, cell := range row {
			if _, ok := index[cell]; ok {
				continue
			}
			index[cell] = len(shared)
			shared = append(shared, cell)
		}
	}

	archive := zip.NewWriter(w)
	files := map[string]string{
		"[Content_Types].xml": contentTypes,
		"_rels/.rels":         rootRels,
		"xl/workbook.xml":     workbookXML(name),
		"xl/_rels/workbook.xml.rels": workbookRels,
		"xl/styles.xml":       styles,
		"xl/sharedStrings.xml": sharedStringsXML(shared),
		"xl/worksheets/sheet1.xml": sheetXML(rows, index),
	}
	for path, body := range files {
		part, err := archive.Create(path)
		if err != nil {
			_ = archive.Close()
			return fmt.Errorf("xlsx part %s: %w", path, err)
		}
		if _, err := io.WriteString(part, body); err != nil {
			_ = archive.Close()
			return fmt.Errorf("xlsx write %s: %w", path, err)
		}
	}
	if err := archive.Close(); err != nil {
		return fmt.Errorf("xlsx close: %w", err)
	}
	return nil
}

func sheetName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "Sheet1"
	}
	replacer := strings.NewReplacer("\\", " ", "/", " ", "*", " ", "?", " ", ":", " ", "[", " ", "]", " ")
	value = replacer.Replace(value)
	if len([]rune(value)) > 31 {
		value = string([]rune(value)[:31])
	}
	return value
}

func escape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}

func workbookXML(name string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" ` +
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<sheets><sheet name="` + escape(name) + `" sheetId="1" r:id="rId1"/></sheets></workbook>`
}

func sharedStringsXML(values []string) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="`)
	builder.WriteString(fmt.Sprintf("%d", len(values)))
	builder.WriteString(`" uniqueCount="`)
	builder.WriteString(fmt.Sprintf("%d", len(values)))
	builder.WriteString(`">`)
	for _, value := range values {
		builder.WriteString(`<si><t>`)
		builder.WriteString(escape(value))
		builder.WriteString(`</t></si>`)
	}
	builder.WriteString(`</sst>`)
	return builder.String()
}

func sheetXML(rows [][]string, index map[string]int) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for rowIndex, row := range rows {
		builder.WriteString(`<row r="`)
		builder.WriteString(fmt.Sprintf("%d", rowIndex+1))
		builder.WriteString(`">`)
		for colIndex, cell := range row {
			ref := cellRef(colIndex, rowIndex)
			builder.WriteString(`<c r="`)
			builder.WriteString(ref)
			builder.WriteString(`" t="s"><v>`)
			builder.WriteString(fmt.Sprintf("%d", index[cell]))
			builder.WriteString(`</v></c>`)
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func cellRef(col, row int) string {
	name := ""
	for col >= 0 {
		name = string(rune('A'+col%26)) + name
		col = col/26 - 1
	}
	return fmt.Sprintf("%s%d", name, row+1)
}

const contentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
	`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
	`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
	`<Default Extension="xml" ContentType="application/xml"/>` +
	`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
	`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>` +
	`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>` +
	`<Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>` +
	`</Types>`

const rootRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
	`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
	`</Relationships>`

const workbookRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
	`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>` +
	`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>` +
	`<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>` +
	`</Relationships>`

const styles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
	`<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
	`<fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts>` +
	`<fills count="1"><fill><patternFill patternType="none"/></fill></fills>` +
	`<borders count="1"><border/></borders>` +
	`<cellStyleXfs count="1"><xf/></cellStyleXfs>` +
	`<cellXfs count="1"><xf/></cellXfs>` +
	`</styleSheet>`
