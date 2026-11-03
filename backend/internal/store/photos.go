package store

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func newID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("id: %w", err)
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}

func detectMedia(header []byte) (string, bool) {
	if len(header) >= 3 && header[0] == 0xff && header[1] == 0xd8 && header[2] == 0xff {
		return "image/jpeg", true
	}
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if len(header) >= len(png) && bytes.Equal(header[:len(png)], png) {
		return "image/png", true
	}
	if len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP" {
		return "image/webp", true
	}
	return "", false
}

func photoRel(deviceID string, photoID string) string {
	return filepath.Join("devices", deviceID, photoID)
}

func photoFull(root string, rel string) (string, error) {
	clean := filepath.Clean(rel)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("photo path")
	}
	full := filepath.Join(root, clean)
	relCheck, err := filepath.Rel(filepath.Clean(root), full)
	if err != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("photo path")
	}
	return full, nil
}

type storedPhoto struct {
	id    string
	rel   string
	size  int64
	media string
}

func writePhoto(root string, deviceID string, r io.Reader) (storedPhoto, error) {
	id, err := newID()
	if err != nil {
		return storedPhoto{}, err
	}
	rel := photoRel(deviceID, id)
	full, err := photoFull(root, rel)
	if err != nil {
		return storedPhoto{}, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return storedPhoto{}, fmt.Errorf("photo directory: %w", err)
	}

	header := make([]byte, 12)
	n, err := io.ReadFull(r, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return storedPhoto{}, fmt.Errorf("read photo: %w", err)
	}
	media, ok := detectMedia(header[:n])
	if !ok {
		return storedPhoto{}, &FieldError{Photo: "type"}
	}

	partial := full + ".partial"
	file, err := os.OpenFile(partial, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return storedPhoto{}, fmt.Errorf("create photo: %w", err)
	}
	written := int64(n)
	_, writeErr := file.Write(header[:n])
	if writeErr == nil {
		var copied int64
		copied, writeErr = io.Copy(file, io.LimitReader(r, maxPhotoBytes-written+1))
		written += copied
	}
	closeErr := file.Close()
	if writeErr != nil {
		os.Remove(partial)
		return storedPhoto{}, fmt.Errorf("write photo: %w", writeErr)
	}
	if closeErr != nil {
		os.Remove(partial)
		return storedPhoto{}, fmt.Errorf("close photo: %w", closeErr)
	}
	if written > maxPhotoBytes {
		os.Remove(partial)
		return storedPhoto{}, &FieldError{Photo: "size"}
	}
	if err := os.Rename(partial, full); err != nil {
		os.Remove(partial)
		return storedPhoto{}, fmt.Errorf("store photo: %w", err)
	}
	return storedPhoto{id: id, rel: rel, size: written, media: media}, nil
}

func removePaths(paths []string) {
	for _, path := range paths {
		if path == "" {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("remove file: %v", err)
		}
	}
}
