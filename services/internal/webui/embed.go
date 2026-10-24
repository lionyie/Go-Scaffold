package webui

import (
	"embed"
	"encoding/json"
	"io/fs"
	"path"
)

//go:embed all:dist
var embedded embed.FS

// AppManifest is the per-app build metadata written by copy-web-dist.mjs.
type AppManifest struct {
	Name     string `json:"name"`
	Base     string `json:"base"`
	Embed    bool   `json:"embed"`
	Fallback bool   `json:"fallback"`
}

// AppFS returns the embedded SPA and its manifest when a build is present.
func AppFS(name string) (fs.FS, AppManifest, bool) {
	sub, err := fs.Sub(embedded, path.Join("dist", name))
	if err != nil {
		return nil, AppManifest{}, false
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, AppManifest{}, false
	}
	manifest := AppManifest{Name: name, Base: "/", Fallback: true}
	if name == "admin" {
		manifest.Base = "/admin/"
	}
	if raw, err := fs.ReadFile(sub, "app.json"); err == nil {
		_ = json.Unmarshal(raw, &manifest)
		if manifest.Base == "" {
			manifest.Base = "/"
		}
	}
	return sub, manifest, true
}
