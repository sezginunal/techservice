package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"techservice/internal/catalog"
	"techservice/internal/store"
)

func homeSummary(database *sql.DB, images catalog.Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := store.LoadSummary(r.Context(), database, time.Now())
		if err != nil {
			log.Printf("home: %v", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		writeJSON(w, http.StatusOK, summaryJSON(summary, images))
	}
}

func summaryJSON(summary store.Summary, images catalog.Index) map[string]any {
	days := make([]map[string]any, 0, len(summary.Days))
	for _, day := range summary.Days {
		days = append(days, map[string]any{"date": day.Date, "count": day.Count})
	}
	byStatus := make([]map[string]any, 0, len(summary.ByStatus))
	for _, item := range summary.ByStatus {
		byStatus = append(byStatus, map[string]any{"status": item.Status, "count": item.Count})
	}
	models := make([]map[string]any, 0, len(summary.Models))
	for _, item := range summary.Models {
		models = append(models, map[string]any{
			"brand": item.Brand,
			"model": item.Model,
			"count": item.Count,
			"image": images.Has(item.Brand, item.Model),
		})
	}
	unchanged := make([]map[string]any, 0, len(summary.Unchanged))
	for _, item := range summary.Unchanged {
		unchanged = append(unchanged, map[string]any{
			"id":           item.ID,
			"brand":        item.Brand,
			"model":        item.Model,
			"customerName": item.CustomerName,
			"status":       item.Status,
			"days":         item.Days,
			"since":        store.FormatTime(item.Since),
		})
	}
	latest := make([]map[string]any, 0, len(summary.Latest))
	for _, item := range summary.Latest {
		latest = append(latest, map[string]any{
			"id":        item.ID,
			"brand":     item.Brand,
			"model":     item.Model,
			"image":     images.Has(item.Brand, item.Model),
			"status":    item.Status,
			"createdAt": store.FormatTime(item.CreatedAt),
		})
	}
	return map[string]any{
		"registered7":   summary.Registered7,
		"registered30":  summary.Registered30,
		"days":          days,
		"inShop":        summary.InShop,
		"delivered":     summary.Delivered,
		"idle7":         summary.Idle7,
		"waitingPart":   summary.WaitingPart,
		"ready":         summary.Ready,
		"byStatus":      byStatus,
		"models":        models,
		"other":         summary.Other,
		"unchanged":     unchanged,
		"unchangedMore": summary.UnchangedMore,
		"latest":        latest,
	}
}
