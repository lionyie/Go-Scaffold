package webui

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/example/scaffold/services/internal/conf"
)

type mount struct {
	name     string
	prefix   string
	fsys     fs.FS
	fallback bool
}

// Handler serves the embedded web and admin single-page applications.
type Handler struct {
	mounts []mount
}

// NewHandler prefers web.dir from config, then the embedded builds. Mount
// prefixes come from each app's app.json.
func NewHandler(bc *conf.Bootstrap) *Handler {
	var mounts []mount
	if dir := bc.GetWeb().GetDir(); dir != "" {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			manifest := AppManifest{Name: "web", Base: "/", Fallback: true}
			if raw, readErr := os.ReadFile(path.Join(dir, "app.json")); readErr == nil {
				_ = json.Unmarshal(raw, &manifest)
				if manifest.Base == "" {
					manifest.Base = "/"
				}
			}
			mounts = append(mounts, mount{name: "web", prefix: manifest.Base, fsys: os.DirFS(dir), fallback: manifest.Fallback})
		}
	}
	for _, name := range []string{"admin", "web"} {
		if fsys, manifest, ok := AppFS(name); ok {
			mounts = append(mounts, mount{
				name:     name,
				prefix:   manifest.Base,
				fsys:     fsys,
				fallback: manifest.Fallback,
			})
		}
	}
	sort.SliceStable(mounts, func(i, j int) bool { return len(mounts[i].prefix) > len(mounts[j].prefix) })
	return &Handler{mounts: mounts}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if target := h.redirectTarget(r.URL.Path); target != "" {
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
		return
	}

	selected := h.pick(r.URL.Path)
	if selected == nil {
		http.Error(w, "web UI is not built", http.StatusNotFound)
		return
	}

	relative := strings.TrimPrefix(r.URL.Path, selected.prefix)
	relative = strings.TrimPrefix(relative, "/")
	clean := path.Clean("/" + relative)
	relative = strings.TrimPrefix(clean, "/")
	if isReserved(relative) {
		http.NotFound(w, r)
		return
	}
	if relative == "" || relative == "." || strings.HasSuffix(r.URL.Path, "/") {
		h.serveIndex(w, r, selected)
		return
	}
	if info, err := fs.Stat(selected.fsys, relative); err == nil && !info.IsDir() {
		if strings.HasPrefix(clean, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.ServeFileFS(w, r, selected.fsys, relative)
		return
	}
	if selected.fallback {
		h.serveIndex(w, r, selected)
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) redirectTarget(requestPath string) string {
	for _, item := range h.mounts {
		if item.prefix == "/" {
			continue
		}
		trimmed := strings.TrimSuffix(item.prefix, "/")
		if requestPath == trimmed {
			return item.prefix
		}
	}
	return ""
}

func (h *Handler) pick(requestPath string) *mount {
	for i := range h.mounts {
		if strings.HasPrefix(requestPath, h.mounts[i].prefix) {
			return &h.mounts[i]
		}
	}
	return nil
}

func (h *Handler) serveIndex(w http.ResponseWriter, r *http.Request, selected *mount) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	file, err := selected.fsys.Open("index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	_, _ = io.Copy(w, file)
}

func isReserved(p string) bool {
	return strings.HasPrefix(p, "api/") ||
		strings.HasPrefix(p, "proxy/") ||
		strings.HasPrefix(p, "healthz") ||
		strings.HasPrefix(p, "readyz") ||
		p == "api" || p == "proxy"
}
