package xlsx

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestWriteContainsSharedStrings(t *testing.T) {
	var buffer bytes.Buffer
	rows := [][]string{
		{"Name", "Phone"},
		{"Ada", "555"},
		{"Ada", "556"},
	}
	if err := Write(&buffer, "Customers", rows); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatal(err)
	}
	parts := map[string]string{}
	for _, file := range reader.File {
		body, err := readZip(file)
		if err != nil {
			t.Fatal(err)
		}
		parts[file.Name] = body
	}
	for _, name := range []string{
		"[Content_Types].xml",
		"xl/workbook.xml",
		"xl/worksheets/sheet1.xml",
		"xl/sharedStrings.xml",
	} {
		if parts[name] == "" {
			t.Fatalf("missing %s", name)
		}
	}
	if !strings.Contains(parts["xl/workbook.xml"], `name="Customers"`) {
		t.Fatalf("sheet name: %s", parts["xl/workbook.xml"])
	}
	if !strings.Contains(parts["xl/sharedStrings.xml"], "Ada") || !strings.Contains(parts["xl/sharedStrings.xml"], "Phone") {
		t.Fatalf("shared strings: %s", parts["xl/sharedStrings.xml"])
	}
	if strings.Count(parts["xl/sharedStrings.xml"], "<si>") != 5 {
		t.Fatalf("unique strings: %s", parts["xl/sharedStrings.xml"])
	}
}

func TestSheetNameSanitizes(t *testing.T) {
	if got := sheetName(`A/B:C*?[]`); got == "" || strings.ContainsAny(got, `\/:*?[]`) {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("ş", 40)
	if got := sheetName(long); len([]rune(got)) != 31 {
		t.Fatalf("got %q", got)
	}
}

func readZip(file *zip.File) (string, error) {
	reader, err := file.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
