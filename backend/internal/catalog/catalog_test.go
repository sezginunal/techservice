package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "devices.json")
	body := `[
		{"brand":" Acme ","models":[" Model 3 ","","S1","s1"]},
		{"brand":"  ","models":["X"]},
		{"brand":"North Wind","models":["A1","T2"]}
	]`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	empty := Load(dir, "  ")
	if len(empty) != 0 {
		t.Fatalf("empty path: %#v", empty)
	}

	loaded := Load(dir, path)
	if len(loaded) != 2 || loaded[0].Brand != "Acme" || len(loaded[0].Models) != 2 || loaded[0].Models[0] != "Model 3" || loaded[0].Models[1] != "S1" {
		t.Fatalf("loaded: %#v", loaded)
	}
	if loaded[1].Brand != "North Wind" || len(loaded[1].Models) != 2 {
		t.Fatalf("second brand: %#v", loaded[1])
	}

	dup := filepath.Join(dir, "dup.json")
	if err := os.WriteFile(dup, []byte(`[{"brand":"Acme","models":["K1"]},{"brand":"acme","models":["K2"]}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	once := Load(dir, dup)
	if len(once) != 1 || len(once[0].Models) != 1 || once[0].Models[0] != "K1" {
		t.Fatalf("duplicate brand: %#v", once)
	}

	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte(`{"brand":"Acme"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if len(Load(dir, bad)) != 0 {
		t.Fatal("invalid json was accepted")
	}

	if len(Load(dir, filepath.Join(dir, "gone.json"))) != 0 {
		t.Fatal("missing file was accepted")
	}

	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if len(Load(dir, filepath.Join(dir, "nested"))) != 0 {
		t.Fatal("directory was accepted")
	}

	outside := filepath.Join(filepath.Dir(dir), "outside.json")
	if err := os.WriteFile(outside, []byte(`[{"brand":"Acme","models":["K1"]}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if len(Load(dir, outside)) != 0 {
		t.Fatal("path outside the brand directory was accepted")
	}

	huge := filepath.Join(dir, "huge.json")
	if err := os.WriteFile(huge, []byte(strings.Repeat(" ", maxSize+1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if len(Load(dir, huge)) != 0 {
		t.Fatal("oversized file was accepted")
	}
}
