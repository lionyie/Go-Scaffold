package plugin

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

var pluginIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,40}$`)

// Manifest is the content of plugins/<id>/plugin.json.
type Manifest struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Frontend    Frontend     `json:"frontend"`
	Backend     Backend      `json:"backend"`
	Permissions []Permission `json:"permissions"`
}

// Frontend describes the browser-side module contributed by a plugin.
type Frontend struct {
	Entry    string   `json:"entry"`
	DevEntry string   `json:"devEntry"`
	Styles   []string `json:"styles"`
	Routes   []Route  `json:"routes"`
}

// Route is a client-side route registered into the host router.
type Route struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Component  string `json:"component"`
	Permission string `json:"permission"`
	Nav        *Nav   `json:"nav,omitempty"`
}

// Nav is an optional sidebar entry for a route.
type Nav struct {
	Label      string `json:"label"`
	Icon       string `json:"icon"`
	Order      int32  `json:"order"`
	Path       string `json:"path"`
	Permission string `json:"permission"`
}

// Backend describes how API calls to /proxy/plugins/<id>/api are served.
type Backend struct {
	Mode                 string `json:"mode"`
	StaticDir            string `json:"staticDir"`
	Upstream             string `json:"upstream"`
	Permission           string `json:"permission"`
	ForwardAuthorization bool   `json:"forwardAuthorization"`
	Timeout              string `json:"timeout"`
}

// Permission is declared by the plugin so the host can seed it.
type Permission struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Plugin couples a validated manifest with its on-disk directory.
type Plugin struct {
	Manifest *Manifest
	Dir      string
}

// Validate normalizes and checks a manifest against its directory name.
func (m *Manifest) Validate(dirName string, dev bool, allowedUpstreams []string) error {
	m.ID = strings.TrimSpace(m.ID)
	if !pluginIDPattern.MatchString(m.ID) {
		return fmt.Errorf("id must match %s", pluginIDPattern.String())
	}
	if m.ID != dirName {
		return fmt.Errorf("id %q must equal directory name %q", m.ID, dirName)
	}
	if !semver.IsValid("v" + strings.TrimPrefix(m.Version, "v")) {
		return fmt.Errorf("version %q is not semver", m.Version)
	}
	if m.Name == "" {
		m.Name = m.ID
	}

	if err := validateRelative(m.Frontend.Entry); err != nil {
		return fmt.Errorf("frontend.entry: %w", err)
	}
	for _, style := range m.Frontend.Styles {
		if err := validateRelative(style); err != nil {
			return fmt.Errorf("frontend.styles: %w", err)
		}
	}
	for i := range m.Frontend.Routes {
		if err := m.Frontend.Routes[i].normalize(m.ID); err != nil {
			return err
		}
	}

	switch m.Backend.Mode {
	case "", "none":
		m.Backend.Mode = "none"
	case "static":
		if m.Backend.StaticDir == "" {
			m.Backend.StaticDir = "server"
		}
		if err := validateRelative(m.Backend.StaticDir); err != nil {
			return fmt.Errorf("backend.staticDir: %w", err)
		}
	case "proxy":
		u, err := url.Parse(m.Backend.Upstream)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("backend.upstream must be an http(s) URL")
		}
		if !dev && !upstreamAllowed(u.Host, allowedUpstreams) {
			return fmt.Errorf("backend.upstream %q is not in plugins.allowed_upstreams", u.Host)
		}
		if m.Backend.Timeout != "" {
			if _, err := time.ParseDuration(m.Backend.Timeout); err != nil {
				return fmt.Errorf("backend.timeout: %w", err)
			}
		}
	default:
		return fmt.Errorf("backend.mode must be one of none, static, proxy")
	}
	return nil
}

func (r *Route) normalize(pluginID string) error {
	if !strings.HasPrefix(r.Path, "/") {
		return fmt.Errorf("route path %q must start with /", r.Path)
	}
	if r.Name == "" {
		r.Name = "plugin:" + pluginID + ":" + strings.ReplaceAll(strings.Trim(r.Path, "/"), "/", "-")
	}
	if r.Nav != nil && r.Nav.Path == "" {
		r.Nav.Path = r.Path
	}
	return nil
}

func validateRelative(p string) error {
	if p == "" {
		return fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(p) || strings.HasPrefix(p, "/") {
		return fmt.Errorf("path %q must be relative", p)
	}
	clean := filepath.ToSlash(filepath.Clean(p))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("path %q escapes the plugin directory", p)
	}
	return nil
}

func upstreamAllowed(host string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(strings.TrimSpace(a), host) {
			return true
		}
	}
	return false
}

// Timeout returns the configured proxy timeout.
func (b Backend) TimeoutDuration() time.Duration {
	if b.Timeout == "" {
		return 30 * time.Second
	}
	d, err := time.ParseDuration(b.Timeout)
	if err != nil || d <= 0 {
		return 30 * time.Second
	}
	return d
}
