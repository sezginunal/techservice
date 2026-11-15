package api

import (
	"log"
	"strings"
)

func NormalizeLang(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "en":
		return "en"
	case "tr":
		return "tr"
	default:
		log.Printf("default language: %q is not en or tr", value)
		return "en"
	}
}
