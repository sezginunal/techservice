package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"techservice/internal/store"
	"techservice/internal/xlsx"
)

const exportPage = 100

type exportCopy struct {
	Sheet        string
	CustomerName string
	Phone        string
	Email        string
	Address      string
	RecordID     string
	Brand        string
	Model        string
	Serial       string
	Customer     string
	Status       string
	RegisteredOn string
	InStock      string
	File         string
}

func exportLabels(lang string) exportCopy {
	if lang == "tr" {
		return exportCopy{
			Sheet:        "Müşteriler",
			CustomerName: "Ad soyad",
			Phone:        "Telefon",
			Email:        "E-posta (isteğe bağlı)",
			Address:      "Adres (isteğe bağlı)",
			RecordID:     "Kimlik",
			Brand:        "Marka",
			Model:        "Model",
			Serial:       "Seri numarası",
			Customer:     "Müşteri",
			Status:       "Durum",
			RegisteredOn: "Kayıt",
			InStock:      "Stok",
			File:         "musteriler",
		}
	}
	return exportCopy{
		Sheet:        "Customers",
		CustomerName: "Name",
		Phone:        "Phone",
		Email:        "Email (optional)",
		Address:      "Address (optional)",
		RecordID:     "ID",
		Brand:        "Brand",
		Model:        "Model",
		Serial:       "Serial number",
		Customer:     "Customer",
		Status:       "Status",
		RegisteredOn: "Registered",
		InStock:      "In stock",
		File:         "customers",
	}
}

func exportServicesLabels(lang string) (sheet, file string) {
	if lang == "tr" {
		return "Servisler", "servisler"
	}
	return "Services", "services"
}

func exportSpareLabels(lang string) (sheet, file string) {
	if lang == "tr" {
		return "Yedek parçalar", "yedek-parcalar"
	}
	return "Spare parts", "spare-parts"
}

func exportCustomers(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := requestLang(r)
		labels := exportLabels(lang)
		query := store.CustomerQuery{Text: r.URL.Query().Get("q"), Limit: exportPage}
		rows := [][]string{{labels.CustomerName, labels.Phone, labels.Email, labels.Address}}
		for {
			page, err := store.ListCustomers(r.Context(), database, query)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			for _, item := range page.Items {
				rows = append(rows, []string{item.Name, item.Phone, item.Email, item.Address})
			}
			if !page.HasMore || len(page.Items) == 0 {
				break
			}
			last := page.Items[len(page.Items)-1]
			query.HasAfter = true
			query.AfterName = last.Name
			query.AfterID = last.ID
		}
		writeExcel(w, labels.File, labels.Sheet, rows)
	}
}

func exportServices(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := requestLang(r)
		labels := exportLabels(lang)
		sheet, file := exportServicesLabels(lang)
		loc := requestLocation(r)
		states, err := store.ListStates(r.Context(), database)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		statusName := map[string]string{}
		for _, item := range states {
			if lang == "tr" {
				statusName[item.Code] = item.NameTr
			} else {
				statusName[item.Code] = item.NameEn
			}
		}
		query := store.DeviceQuery{
			Text:   r.URL.Query().Get("q"),
			Status: r.URL.Query().Get("status"),
			Limit:  exportPage,
		}
		rows := [][]string{{
			labels.RecordID,
			labels.Brand,
			labels.Model,
			labels.Serial,
			labels.Customer,
			labels.Status,
			labels.RegisteredOn,
		}}
		for {
			page, err := store.ListDevices(r.Context(), database, query)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			for _, item := range page.Items {
				status := statusName[item.Status]
				if status == "" {
					status = item.Status
				}
				rows = append(rows, []string{
					item.ID,
					item.Brand,
					item.Model,
					item.Serial,
					item.CustomerName,
					status,
					formatWhen(item.CreatedAt, lang, loc),
				})
			}
			if !page.HasMore || len(page.Items) == 0 {
				break
			}
			last := page.Items[len(page.Items)-1]
			query.HasAfter = true
			query.AfterTime = store.FormatTime(last.CreatedAt)
			query.AfterID = last.ID
		}
		writeExcel(w, file, sheet, rows)
	}
}

func exportSpareParts(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := requestLang(r)
		labels := exportLabels(lang)
		sheet, file := exportSpareLabels(lang)
		query := store.SpareListQuery{Text: r.URL.Query().Get("q"), Limit: exportPage}
		rows := [][]string{{labels.Brand, labels.Model, labels.InStock}}
		for {
			page, err := store.ListSpareParts(r.Context(), database, query)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			for _, item := range page.Items {
				rows = append(rows, []string{item.Brand, item.Model, strconv.Itoa(item.OnHand)})
			}
			if !page.HasMore || len(page.Items) == 0 {
				break
			}
			last := page.Items[len(page.Items)-1]
			query.HasAfter = true
			query.AfterBrand = last.Brand
			query.AfterModel = last.Model
			query.AfterID = last.ID
		}
		writeExcel(w, file, sheet, rows)
	}
}

func writeExcel(w http.ResponseWriter, fileName, sheet string, rows [][]string) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, fileName))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if err := xlsx.Write(w, sheet, rows); err != nil {
		log.Printf("excel export: %v", err)
	}
}

func requestLang(r *http.Request) string {
	return NormalizeLang(r.URL.Query().Get("lang"))
}

func requestLocation(r *http.Request) *time.Location {
	name := strings.TrimSpace(r.URL.Query().Get("tz"))
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

var monthsTR = [...]string{"Oca", "Şub", "Mar", "Nis", "May", "Haz", "Tem", "Ağu", "Eyl", "Eki", "Kas", "Ara"}

func formatWhen(value time.Time, lang string, loc *time.Location) string {
	local := value.In(loc)
	if lang == "tr" {
		return fmt.Sprintf("%d %s %d %02d:%02d", local.Day(), monthsTR[local.Month()-1], local.Year(), local.Hour(), local.Minute())
	}
	return local.Format("2 Jan 2006, 15:04")
}
