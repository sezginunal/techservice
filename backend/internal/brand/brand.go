package brand

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	Dir         = "/brand"
	maxLogoSize = 2 << 20
	maxName     = 200
)

type Brand struct {
	Name  string
	media string
	body  []byte
}

func (b Brand) Logo() (body []byte, media string, ok bool) {
	if b.media == "" {
		return nil, "", false
	}
	return b.body, b.media, true
}

func Load(dir, name, path string) Brand {
	loaded := Brand{Name: cleanName(name)}
	if strings.TrimSpace(path) == "" {
		return loaded
	}
	body, media, err := readLogo(dir, path)
	if err != nil {
		log.Printf("logo: %v", err)
		return loaded
	}
	loaded.body = body
	loaded.media = media
	return loaded
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if utf8.RuneCountInString(name) > maxName {
		log.Printf("company name: longer than %d characters", maxName)
		return ""
	}
	return name
}

func readLogo(dir, path string) ([]byte, string, error) {
	full, err := inside(dir, path)
	if err != nil {
		return nil, "", err
	}
	file, err := os.Open(full)
	if err != nil {
		return nil, "", fmt.Errorf("open %s: %w", full, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("stat %s: %w", full, err)
	}
	if !info.Mode().IsRegular() {
		return nil, "", fmt.Errorf("%s is not a file", full)
	}
	body, err := io.ReadAll(io.LimitReader(file, maxLogoSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", full, err)
	}
	if len(body) > maxLogoSize {
		return nil, "", fmt.Errorf("%s is larger than %d bytes", full, maxLogoSize)
	}
	media, ok := mediaType(body)
	if !ok {
		return nil, "", fmt.Errorf("%s is not a JPEG, PNG, WebP, or SVG", full)
	}
	return body, media, nil
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

func mediaType(body []byte) (string, bool) {
	if len(body) >= 3 && body[0] == 0xff && body[1] == 0xd8 && body[2] == 0xff {
		return "image/jpeg", true
	}
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if len(body) >= len(png) && bytes.Equal(body[:len(png)], png) {
		return "image/png", true
	}
	if len(body) >= 12 && string(body[0:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return "image/webp", true
	}
	if svg(body) {
		return "image/svg+xml", true
	}
	return "", false
}

func svg(body []byte) bool {
	if len(body) == 0 || bytes.IndexByte(body, 0) >= 0 {
		return false
	}
	sample := body
	if len(sample) > 1024 {
		sample = sample[:1024]
	}
	text := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(string(sample), "\uFEFF")))
	if strings.HasPrefix(text, "<svg") {
		return true
	}
	return strings.HasPrefix(text, "<?xml") && strings.Contains(text, "<svg")
}
