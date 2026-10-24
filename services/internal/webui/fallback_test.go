package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testHandler() *Handler {
	admin := fstest.MapFS{
		"index.html":  &fstest.MapFile{Data: []byte("admin-index")},
		"assets/x.js": &fstest.MapFile{Data: []byte("admin-js")},
		"app.json":    &fstest.MapFile{Data: []byte(`{"base":"/admin/"}`)},
	}
	web := fstest.MapFS{
		"index.html":  &fstest.MapFile{Data: []byte("web-index")},
		"assets/x.js": &fstest.MapFile{Data: []byte("web-js")},
	}
	return &Handler{mounts: []mount{
		{name: "admin", prefix: "/admin/", fsys: admin, fallback: true},
		{name: "web", prefix: "/", fsys: web, fallback: true},
	}}
}

func TestSPAFallback(t *testing.T) {
	handler := testHandler()
	cases := []struct {
		path   string
		status int
		body   string
	}{
		{"/", http.StatusOK, "web-index"},
		{"/unknown", http.StatusOK, "web-index"},
		{"/admin/", http.StatusOK, "admin-index"},
		{"/admin/users", http.StatusOK, "admin-index"},
		{"/admin/assets/x.js", http.StatusOK, "admin-js"},
		{"/assets/x.js", http.StatusOK, "web-js"},
		{"/api/v1/plugins", http.StatusNotFound, ""},
		{"/proxy/plugins/x/frontend/index.js", http.StatusNotFound, ""},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			handler.ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.status)
			}
			body, _ := io.ReadAll(recorder.Result().Body)
			if tc.body != "" && !strings.Contains(string(body), tc.body) {
				t.Fatalf("body = %q, want contains %q", string(body), tc.body)
			}
		})
	}
}

func TestAdminRedirect(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := testHandler()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if recorder.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusPermanentRedirect)
	}
	if location := recorder.Header().Get("Location"); location != "/admin/" {
		t.Fatalf("location = %q", location)
	}
}
