package swagger

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/scaffold/services/internal/conf"
)

func TestAllowed(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *conf.Swagger
		appEnv string
		want   bool
	}{
		{name: "disabled config", cfg: &conf.Swagger{Enabled: false}, appEnv: "development", want: false},
		{name: "nil config", cfg: nil, appEnv: "development", want: false},
		{name: "enabled development", cfg: &conf.Swagger{Enabled: true}, appEnv: "development", want: true},
		{name: "enabled staging", cfg: &conf.Swagger{Enabled: true}, appEnv: "staging", want: true},
		{name: "enabled empty env", cfg: &conf.Swagger{Enabled: true}, appEnv: "", want: true},
		{name: "production hard off", cfg: &conf.Swagger{Enabled: true}, appEnv: "production", want: false},
		{name: "production case and spaces", cfg: &conf.Swagger{Enabled: true}, appEnv: "  Production ", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Allowed(tt.cfg, tt.appEnv); got != tt.want {
				t.Fatalf("Allowed(%v, %q) = %v, want %v", tt.cfg, tt.appEnv, got, tt.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	handler := NewHandler()

	t.Run("redirects base path", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/swagger", nil))

		if rec.Code != http.StatusMovedPermanently {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusMovedPermanently)
		}
		if location := rec.Header().Get("Location"); location != "/swagger/" {
			t.Fatalf("Location = %q, want %q", location, "/swagger/")
		}
	})

	t.Run("serves openapi spec", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/yaml") {
			t.Fatalf("Content-Type = %q, want application/yaml", ct)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "openapi: 3.0.3") {
			t.Fatalf("spec does not declare openapi 3.0.3")
		}
		if !strings.Contains(body, "title: Scaffold API") {
			t.Fatalf("spec does not contain the Scaffold API title")
		}
	})

	t.Run("serves swagger ui", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/swagger/", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Fatalf("Content-Type = %q, want text/html", ct)
		}
		if rec.Body.Len() == 0 {
			t.Fatal("expected a non-empty Swagger UI page")
		}
	})
}
