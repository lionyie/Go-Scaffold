package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/conf"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/rbac"
)

const proxyPrefix = "/proxy/plugins/"

// Registry holds the validated plugins loaded from disk.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
	dev     bool
	checker rbac.Checker
	tokens  *authpkg.TokenManager
	log     *slog.Logger
}

// NewRegistry loads plugins from the configured directory.
func NewRegistry(
	bc *conf.Bootstrap,
	checker rbac.Checker,
	tokens *authpkg.TokenManager,
	roles biz.RoleRepo,
	log *slog.Logger,
) (*Registry, func(), error) {
	cfg := bc.GetPlugins()
	plugins, warnings, err := Load(cfg.GetDir(), cfg.GetDev(), cfg.GetAllowedUpstreams(), cfg.GetStrict())
	if err != nil {
		return nil, nil, err
	}
	for _, warning := range warnings {
		log.Warn("skipping invalid plugin", "err", warning)
	}

	registry := &Registry{
		plugins: make(map[string]*Plugin, len(plugins)),
		dev:     cfg.GetDev(),
		checker: checker,
		tokens:  tokens,
		log:     log,
	}
	for _, p := range plugins {
		registry.plugins[p.Manifest.ID] = p
		log.Info("loaded plugin", "id", p.Manifest.ID, "version", p.Manifest.Version)
	}
	if err := registry.syncPermissions(context.Background(), roles); err != nil {
		return nil, nil, err
	}
	return registry, func() {}, nil
}

// syncPermissions makes plugin-declared permissions grantable through the role
// API even though the built-in catalog lives in code.
func (r *Registry) syncPermissions(ctx context.Context, roles biz.RoleRepo) error {
	var permissions []*biz.Permission
	for _, p := range r.plugins {
		for _, perm := range p.Manifest.Permissions {
			permissions = append(permissions, &biz.Permission{
				Code:     perm.Code,
				Name:     perm.Name,
				Resource: "plugin",
				Action:   p.Manifest.ID,
			})
		}
	}
	if len(permissions) == 0 {
		return nil
	}
	return roles.UpsertPermissions(ctx, permissions)
}

// Handler serves plugin assets and API proxies.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(r.serveHTTP)
}

func (r *Registry) serveHTTP(w http.ResponseWriter, rq *http.Request) {
	rest := strings.TrimPrefix(rq.URL.Path, proxyPrefix)
	parts := strings.SplitN(rest, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		writeJSONError(w, http.StatusNotFound, "PLUGIN_NOT_FOUND", "plugin route not found")
		return
	}
	pluginID, kind := parts[0], parts[1]
	tail := ""
	if len(parts) == 3 {
		tail = parts[2]
	}

	p, ok := r.get(pluginID)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "PLUGIN_NOT_FOUND", "plugin not found")
		return
	}

	switch kind {
	case "frontend":
		w.Header().Set("Cache-Control", "public, max-age=3600")
		r.serveFile(w, rq, filepath.Join(p.Dir, "dist"), tail)
	case "api":
		r.serveAPI(w, rq, p, tail)
	default:
		writeJSONError(w, http.StatusNotFound, "PLUGIN_NOT_FOUND", "unknown plugin endpoint")
	}
}

func (r *Registry) serveAPI(w http.ResponseWriter, rq *http.Request, p *Plugin, tail string) {
	principal, ok := r.authenticate(rq)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	if principal.MustChangePassword {
		writeJSONError(w, http.StatusForbidden, "MUST_CHANGE_PASSWORD", "password change required")
		return
	}
	if !principal.IsAdmin && p.Manifest.Backend.Permission != "" {
		allowed, err := r.checker.HasGlobal(rq.Context(), principal.UserID, p.Manifest.Backend.Permission)
		if err != nil || !allowed {
			writeJSONError(w, http.StatusForbidden, "FORBIDDEN", "permission denied")
			return
		}
	}

	switch p.Manifest.Backend.Mode {
	case "static":
		w.Header().Set("Cache-Control", "no-store")
		r.serveFile(w, rq, filepath.Join(p.Dir, p.Manifest.Backend.StaticDir), tail)
	case "proxy":
		r.proxy(w, rq, p, principal, tail)
	default:
		writeJSONError(w, http.StatusNotFound, "PLUGIN_NOT_FOUND", "plugin has no backend")
	}
}

func (r *Registry) proxy(w http.ResponseWriter, rq *http.Request, p *Plugin, principal *authpkg.Principal, tail string) {
	target, err := url.Parse(p.Manifest.Backend.Upstream)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "PLUGIN_UPSTREAM", "invalid upstream")
		return
	}
	pluginID := p.Manifest.ID
	userID := principal.UserID
	username := principal.Username
	timeout := p.Manifest.Backend.TimeoutDuration()

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.URL.Path = joinURLPath(target.Path, tail)
			pr.Out.Host = target.Host
			pr.Out.Header.Set("X-Scaffold-Plugin-Id", pluginID)
			pr.Out.Header.Set("X-Scaffold-User-Id", userID)
			pr.Out.Header.Set("X-Scaffold-Username", username)
			if !p.Manifest.Backend.ForwardAuthorization {
				pr.Out.Header.Del("Authorization")
			}
		},
		ErrorHandler: func(rw http.ResponseWriter, _ *http.Request, err error) {
			r.log.Warn("plugin upstream failed", "plugin", pluginID, "err", err)
			writeJSONError(rw, http.StatusBadGateway, "PLUGIN_UPSTREAM", "plugin upstream unavailable")
		},
	}

	ctx, cancel := context.WithTimeout(rq.Context(), timeout)
	defer cancel()
	proxy.ServeHTTP(w, rq.WithContext(ctx))
}

func (r *Registry) serveFile(w http.ResponseWriter, rq *http.Request, base, rel string) {
	if rel == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
		return
	}
	clean := path.Clean("/" + rel)
	full := filepath.Join(base, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
	absBase, err := filepath.Abs(base)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "invalid plugin directory")
		return
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "invalid plugin path")
		return
	}
	if absFull != absBase && !strings.HasPrefix(absFull, absBase+string(os.PathSeparator)) {
		writeJSONError(w, http.StatusForbidden, "FORBIDDEN", "path escapes the plugin directory")
		return
	}
	if info, err := os.Stat(absFull); err != nil || info.IsDir() {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
		return
	}
	http.ServeFile(w, rq, absFull)
}

func (r *Registry) authenticate(rq *http.Request) (*authpkg.Principal, bool) {
	token := authpkg.Bearer(rq.Header.Get("Authorization"))
	if token == "" {
		return nil, false
	}
	claims, err := r.tokens.ParseAccess(token)
	if err != nil {
		return nil, false
	}
	return &authpkg.Principal{
		UserID:             claims.UserID,
		Username:           claims.Username,
		SessionID:          claims.SessionID,
		IsAdmin:            claims.Admin,
		MustChangePassword: claims.MustChangePassword,
	}, true
}

func (r *Registry) get(id string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[id]
	return p, ok
}

// ListPlugins returns the manifests visible to the caller.
func (r *Registry) ListPlugins(ctx context.Context, userID string, admin bool) ([]*biz.PluginInfo, error) {
	r.mu.RLock()
	plugins := make([]*Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	r.mu.RUnlock()

	out := make([]*biz.PluginInfo, 0, len(plugins))
	for _, p := range plugins {
		visible, err := r.visible(ctx, p, userID, admin)
		if err != nil {
			return nil, err
		}
		if visible {
			out = append(out, r.toInfo(p))
		}
	}
	return out, nil
}

// GetPlugin returns a single plugin regardless of the caller's permissions.
func (r *Registry) GetPlugin(_ context.Context, id string) (*biz.PluginInfo, error) {
	p, ok := r.get(id)
	if !ok {
		return nil, biz.ErrPluginNotFound()
	}
	return r.toInfo(p), nil
}

func (r *Registry) visible(ctx context.Context, p *Plugin, userID string, admin bool) (bool, error) {
	if admin || len(p.Manifest.Permissions) == 0 {
		return true, nil
	}
	for _, perm := range p.Manifest.Permissions {
		allowed, err := r.checker.HasGlobal(ctx, userID, perm.Code)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

func (r *Registry) toInfo(p *Plugin) *biz.PluginInfo {
	m := p.Manifest
	// Manifest entry/styles are paths inside the plugin's dist/ directory.
	// The public URL always goes through the frontend asset route.
	base := proxyPrefix + m.ID + "/frontend/"
	routes := make([]biz.PluginRoute, 0, len(m.Frontend.Routes))
	navs := make([]biz.PluginNav, 0, len(m.Frontend.Routes))
	for _, route := range m.Frontend.Routes {
		routes = append(routes, biz.PluginRoute{
			Path:       route.Path,
			Name:       route.Name,
			Component:  route.Component,
			Permission: route.Permission,
		})
		if route.Nav != nil {
			navs = append(navs, biz.PluginNav{
				Label:      route.Nav.Label,
				Icon:       route.Nav.Icon,
				Order:      route.Nav.Order,
				Path:       route.Nav.Path,
				Permission: route.Nav.Permission,
			})
		}
	}
	permissions := make([]string, 0, len(m.Permissions))
	for _, perm := range m.Permissions {
		permissions = append(permissions, perm.Code)
	}
	styles := make([]string, 0, len(m.Frontend.Styles))
	for _, style := range m.Frontend.Styles {
		styles = append(styles, base+strings.TrimPrefix(style, "/"))
	}

	entry := base + strings.TrimPrefix(m.Frontend.Entry, "/")
	if r.dev && m.Frontend.DevEntry != "" {
		entry = m.Frontend.DevEntry
	}
	return &biz.PluginInfo{
		ID:          m.ID,
		Name:        m.Name,
		Version:     m.Version,
		Description: m.Description,
		Icon:        m.Icon,
		EntryURL:    entry,
		Styles:      styles,
		Routes:      routes,
		Navigations: navs,
		Permissions: permissions,
	}
}

func joinURLPath(base, tail string) string {
	if tail == "" {
		return base
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(tail, "/")
}

func writeJSONError(w http.ResponseWriter, status int, reason, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    status,
		"reason":  reason,
		"message": message,
	})
}

// String is useful for tests and logs.
func (r *Registry) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return fmt.Sprintf("plugins=%d", len(r.plugins))
}
