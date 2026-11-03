package catalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const maxSize = 1 << 20

type Brand struct {
	Brand  string   `json:"brand"`
	Models []string `json:"models"`
}

func Load(dir, path string) []Brand {
	if strings.TrimSpace(path) == "" {
		return []Brand{}
	}
	loaded, err := read(dir, path)
	if err != nil {
		log.Printf("device catalog: %v", err)
		return []Brand{}
	}
	return loaded
}

func read(dir, path string) ([]Brand, error) {
	full, err := inside(dir, path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(full)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", full, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", full, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a file", full)
	}
	body, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", full, err)
	}
	if len(body) > maxSize {
		return nil, fmt.Errorf("%s is larger than %d bytes", full, maxSize)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var raw []Brand
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", full, err)
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("parse %s: extra value", full)
		}
		return nil, fmt.Errorf("parse %s: %w", full, err)
	}
	return clean(raw), nil
}

func clean(raw []Brand) []Brand {
	loaded := make([]Brand, 0, len(raw))
	seen := map[string]bool{}
	for _, item := range raw {
		name := strings.TrimSpace(item.Brand)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			log.Printf("device catalog: duplicate brand %q", name)
			continue
		}
		seen[key] = true
		models := make([]string, 0, len(item.Models))
		modelSeen := map[string]bool{}
		for _, model := range item.Models {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			modelKey := strings.ToLower(model)
			if modelSeen[modelKey] {
				continue
			}
			modelSeen[modelKey] = true
			models = append(models, model)
		}
		loaded = append(loaded, Brand{Brand: name, Models: models})
	}
	return loaded
}

func inside(dir, path string) (string, error) {
	root := filepath.Clean(dir)
	full := filepath.Clean(path)
	rel, err := filepath.Rel(root, full)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s is outside %s", path, dir)
	}
	return full, nil
}
