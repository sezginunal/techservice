package catalog

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const maxImage = 2 << 20

type Index struct {
	Dir    string
	Brands []Brand
}

func (index Index) Has(brand, model string) bool {
	_, _, ok := index.locate(brand, model)
	return ok
}

func (index Index) Open(brand, model string) ([]byte, string, bool) {
	path, media, ok := index.locate(brand, model)
	if !ok {
		return nil, "", false
	}
	file, err := os.Open(path)
	if err != nil {
		log.Printf("device catalog image: open %s: %v", path, err)
		return nil, "", false
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, maxImage+1))
	if err != nil {
		log.Printf("device catalog image: read %s: %v", path, err)
		return nil, "", false
	}
	if len(body) > maxImage {
		log.Printf("device catalog image: %s is larger than %d bytes", path, maxImage)
		return nil, "", false
	}
	return body, media, true
}

func (index Index) locate(brand, model string) (string, string, bool) {
	canonBrand, canonModel, ok := match(index.Brands, brand, model)
	if !ok || strings.TrimSpace(index.Dir) == "" {
		return "", "", false
	}
	brandSlug := Slug(canonBrand)
	modelSlug := Slug(canonModel)
	if brandSlug == "" || modelSlug == "" {
		return "", "", false
	}
	root := filepath.Join(index.Dir, "devices")
	for _, ext := range []string{".webp", ".png", ".jpg", ".jpeg"} {
		full, err := inside(root, filepath.Join(root, brandSlug, modelSlug+ext))
		if err != nil {
			log.Printf("device catalog image: %v", err)
			continue
		}
		media, err := sniff(full)
		if err != nil {
			if !errors.Is(err, errMissing) {
				log.Printf("device catalog image: %v", err)
			}
			continue
		}
		return full, media, true
	}
	return "", "", false
}

func match(brands []Brand, brand, model string) (string, string, bool) {
	brand = strings.TrimSpace(brand)
	model = strings.TrimSpace(model)
	for _, item := range brands {
		if !strings.EqualFold(item.Brand, brand) {
			continue
		}
		for _, name := range item.Models {
			if strings.EqualFold(name, model) {
				return item.Brand, name, true
			}
		}
	}
	return "", "", false
}

func Slug(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "+", " plus ")
	var b strings.Builder
	dash := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

var errMissing = errors.New("missing")

func sniff(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errMissing
		}
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a file", path)
	}
	if info.Size() > maxImage {
		return "", fmt.Errorf("%s is larger than %d bytes", path, maxImage)
	}
	header := make([]byte, 32)
	n, err := io.ReadFull(file, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	media, ok := imageType(header[:n])
	if !ok {
		return "", fmt.Errorf("%s is not a JPEG, PNG, or WebP", path)
	}
	return media, nil
}

func imageType(body []byte) (string, bool) {
	if len(body) >= 3 && body[0] == 0xff && body[1] == 0xd8 && body[2] == 0xff {
		return "image/jpeg", true
	}
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if len(body) >= len(png) && string(body[:len(png)]) == string(png) {
		return "image/png", true
	}
	if len(body) >= 12 && string(body[0:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return "image/webp", true
	}
	return "", false
}
