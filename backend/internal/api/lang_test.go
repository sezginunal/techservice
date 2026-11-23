package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"techservice/internal/brand"
	"techservice/internal/catalog"
)

func TestNormalizeLang(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "", want: "en"},
		{in: "en", want: "en"},
		{in: " EN ", want: "en"},
		{in: "tr", want: "tr"},
		{in: "TR", want: "tr"},
		{in: "de", want: "en"},
	}
	for _, item := range cases {
		if got := NormalizeLang(item.in); got != item.want {
			t.Fatalf("%q: got %s", item.in, got)
		}
	}
}

func TestBrandIncludesLanguage(t *testing.T) {
	handler := New(nil, t.TempDir(), brand.Brand{Name: "North"}, catalog.Files{}, "tr")
	request := httptest.NewRequest(http.MethodGet, "/api/brand", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d", recorder.Code)
	}
	var body struct {
		Name string `json:"name"`
		Logo bool   `json:"logo"`
		Lang string `json:"lang"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Name != "North" || body.Logo || body.Lang != "tr" {
		t.Fatalf("%#v", body)
	}
}
