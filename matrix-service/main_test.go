package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicAPI(t *testing.T) {
	statsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/statistics" {
			t.Errorf("unexpected statistics path %s", r.URL.Path)
		}
		var request statisticsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if len(request.Matrices) == 0 {
			t.Error("statistics request has no matrices")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"min":1,"max":6,"sum":21,"average":3.5,"hasDiagonalMatrix":false}`)
	}))
	defer statsServer.Close()
	app := newApp(statsServer.URL+"/api/v1/statistics", statsServer.Client())
	for _, test := range []struct {
		name   string
		path   string
		fields []string
	}{
		{"rotate", "/api/v1/matrices/rotate", []string{"rotated", "statistics"}},
		{"QR", "/api/v1/matrices/qr", []string{"q", "r", "statistics"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(`{"matrix":[[1,2,3],[4,5,6]]}`))
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", response.StatusCode)
			}
			var body map[string]json.RawMessage
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			for _, field := range test.fields {
				if _, ok := body[field]; !ok {
					t.Errorf("response has no %q: %v", field, body)
				}
			}
		})
	}
}

func TestDocs(t *testing.T) {
	app := newApp("http://example.invalid", http.DefaultClient)
	request := httptest.NewRequest(http.MethodGet, "/docs", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("content type = %q", contentType)
	}
}
