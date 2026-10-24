package plugin

import "testing"

func TestToInfoAssetURLs(t *testing.T) {
	p := &Plugin{Manifest: &Manifest{
		ID:      "hello-panel",
		Name:    "Hello Panel",
		Version: "0.1.0",
		Frontend: Frontend{
			Entry:  "index.js",
			Styles: []string{"style.css"},
			Routes: []Route{{
				Path:      "/plugins/hello-panel",
				Name:      "plugin-hello-panel",
				Component: "HelloView",
				Nav: &Nav{
					Label: "Hello plugin",
					Icon:  "mdi-hand-wave",
					Order: 60,
					Path:  "/plugins/hello-panel",
				},
			}},
		},
	}}

	registry := &Registry{dev: false}
	info := registry.toInfo(p)

	if info.EntryURL != "/proxy/plugins/hello-panel/frontend/index.js" {
		t.Fatalf("entry url = %q", info.EntryURL)
	}
	if len(info.Styles) != 1 || info.Styles[0] != "/proxy/plugins/hello-panel/frontend/style.css" {
		t.Fatalf("styles = %#v", info.Styles)
	}
	if len(info.Navigations) != 1 || info.Navigations[0].Icon != "mdi-hand-wave" {
		t.Fatalf("navigations = %#v", info.Navigations)
	}
}

func TestToInfoDevEntry(t *testing.T) {
	p := &Plugin{Manifest: &Manifest{
		ID:      "hello-panel",
		Name:    "Hello Panel",
		Version: "0.1.0",
		Frontend: Frontend{
			Entry:    "index.js",
			DevEntry: "http://localhost:5174/src/index.ts",
			Styles:   []string{"style.css"},
		},
	}}

	registry := &Registry{dev: true}
	info := registry.toInfo(p)
	if info.EntryURL != "http://localhost:5174/src/index.ts" {
		t.Fatalf("dev entry url = %q", info.EntryURL)
	}
	if info.Styles[0] != "/proxy/plugins/hello-panel/frontend/style.css" {
		t.Fatalf("dev styles = %#v", info.Styles)
	}
}
