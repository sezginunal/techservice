package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Acme":    "acme",
		"S1":         "s1",
		"North Wind":   "north-wind",
		"Mini 2":     "mini-2",
		"XL+":         "xl-plus",
		"2.4":         "2-4",
		"2.0 Model T":   "2-0-model-t",
		"ACME-PRO":   "acme-pro",
		"  Model 3  ": "model-3",
	}
	for name, want := range cases {
		if got := Slug(name); got != want {
			t.Fatalf("Slug(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestImage(t *testing.T) {
	dir := t.TempDir()
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00}
	webp := append([]byte("RIFF"), append(make([]byte, 4), []byte("WEBP")...)...)
	brands := []Brand{{Brand: "Acme", Models: []string{"S1", "Model 3"}}, {Brand: "Helios", Models: []string{"XL+"}}}
	index := Index{Dir: dir, Brands: brands}

	if index.Has("Acme", "S1") {
		t.Fatal("missing file was reported")
	}
	if index.Has("Acme", "Other") {
		t.Fatal("unknown model was reported")
	}

	modelDir := filepath.Join(dir, "devices", "acme")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "s1.png"), png, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "s1.webp"), webp, 0o644); err != nil {
		t.Fatal(err)
	}
	body, media, ok := index.Open("  acme ", "s1")
	if !ok || media != "image/webp" || string(body[:4]) != "RIFF" {
		t.Fatalf("webp: %s %v %d", media, ok, len(body))
	}

	if err := os.WriteFile(filepath.Join(modelDir, "model-3.png"), []byte("not an image!!!!"), 0o644); err != nil {
		t.Fatal(err)
	}
	if index.Has("Acme", "Model 3") {
		t.Fatal("text file was accepted")
	}

	plusDir := filepath.Join(dir, "devices", "helios")
	if err := os.MkdirAll(plusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	jpeg := []byte{0xff, 0xd8, 0xff, 0x00}
	if err := os.WriteFile(filepath.Join(plusDir, "xl-plus.jpg"), jpeg, 0o644); err != nil {
		t.Fatal(err)
	}
	_, media, ok = index.Open("Helios", "XL+")
	if !ok || media != "image/jpeg" {
		t.Fatalf("jpeg: %s %v", media, ok)
	}

	huge := filepath.Join(modelDir, "s1.webp")
	if err := os.WriteFile(huge, append([]byte("RIFF"), append(make([]byte, 4), append([]byte("WEBP"), make([]byte, maxImage)...)...)...), 0o644); err != nil {
		t.Fatal(err)
	}
	body, media, ok = index.Open("Acme", "S1")
	if !ok || media != "image/png" {
		t.Fatalf("oversized webp should fall through to png: %s %v %d", media, ok, len(body))
	}
}
