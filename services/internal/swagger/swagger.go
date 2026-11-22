package swagger

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/swaggest/swgui/v5emb"

	"github.com/example/scaffold/services/internal/conf"
)

//go:embed generated/openapi.yaml
var spec []byte

// Allowed reports whether the Swagger UI may be served. A production
// environment always disables it, regardless of the config flag.
func Allowed(cfg *conf.Swagger, appEnv string) bool {
	if cfg == nil || !cfg.GetEnabled() {
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(appEnv), "production")
}

// NewHandler returns the Swagger UI and the embedded OpenAPI document.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(spec)
	})
	mux.Handle("/swagger/", v5emb.New("Scaffold API", "/swagger/openapi.yaml", "/swagger/"))
	return mux
}
