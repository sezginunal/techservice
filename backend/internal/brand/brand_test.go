package brand

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	png := append([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}, []byte("png")...)
	logo := filepath.Join(dir, "logo.png")
	if err := os.WriteFile(logo, png, 0o644); err != nil {
		t.Fatal(err)
	}
	svgPath := filepath.Join(dir, "mark.svg")
	if err := os.WriteFile(svgPath, []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	empty := Load(dir, "  ", "")
	if empty.Name != "" || logoOK(empty) {
		t.Fatalf("empty: %#v", empty)
	}

	named := Load(dir, "  North Shop  ", "")
	if named.Name != "North Shop" || logoOK(named) {
		t.Fatalf("name: %#v", named)
	}

	long := Load(dir, strings.Repeat("a", 201), logo)
	if long.Name != "" || !logoOK(long) {
		t.Fatalf("long name: %#v", long.Name)
	}

	withLogo := Load(dir, "North Shop", logo)
	body, media, ok := withLogo.Logo()
	if !ok || media != "image/png" || !bytes.Equal(body, png) {
		t.Fatalf("png: %s %v", media, ok)
	}

	svgLogo := Load(dir, "", svgPath)
	_, media, ok = svgLogo.Logo()
	if !ok || media != "image/svg+xml" {
		t.Fatalf("svg: %s %v", media, ok)
	}

	missing := Load(dir, "North Shop", filepath.Join(dir, "gone.png"))
	if missing.Name != "North Shop" || logoOK(missing) {
		t.Fatal("missing file changed the name or kept a logo")
	}

	text := Load(dir, "North Shop", filepath.Join(dir, "note.txt"))
	if logoOK(text) {
		t.Fatal("text file was accepted")
	}

	outside := filepath.Join(filepath.Dir(dir), "outside.png")
	if err := os.WriteFile(outside, png, 0o644); err != nil {
		t.Fatal(err)
	}
	escaped := Load(dir, "North Shop", outside)
	if logoOK(escaped) {
		t.Fatal("path outside the brand directory was accepted")
	}
}

func logoOK(b Brand) bool {
	_, _, ok := b.Logo()
	return ok
}
