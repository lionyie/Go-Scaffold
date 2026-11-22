package plugin

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/conf"
)

const (
	maxArchiveBytes = 10 << 20
	maxExtractBytes = 50 << 20
	maxArchiveFile  = 10 << 20
	maxArchiveFiles = 5000
	gitCloneTimeout = 60 * time.Second
)

// Installer installs and removes plugins from a writable directory.
type Installer struct {
	installDir       string
	allowedRepos     []string
	allowedUpstreams []string
	registry         *Registry
	log              *slog.Logger
}

type installMeta struct {
	Source      string    `json:"source"`
	URL         string    `json:"url,omitempty"`
	Ref         string    `json:"ref,omitempty"`
	Commit      string    `json:"commit,omitempty"`
	InstalledAt time.Time `json:"installedAt"`
}

// NewInstaller builds the installer and ensures the install directory exists.
func NewInstaller(bc *conf.Bootstrap, registry *Registry, log *slog.Logger) (*Installer, error) {
	cfg := bc.GetPlugins()
	dir := strings.TrimSpace(cfg.GetInstallDir())
	if dir == "" {
		dir = filepath.Join("..", "data", "plugins")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create plugin install dir: %w", err)
	}
	return &Installer{
		installDir:       dir,
		allowedRepos:     cfg.GetAllowedRepos(),
		allowedUpstreams: cfg.GetAllowedUpstreams(),
		registry:         registry,
		log:              log,
	}, nil
}

// LoadInstalled registers every plugin found under the install directory.
func (i *Installer) LoadInstalled() ([]*Plugin, error) {
	plugins, warnings, err := LoadWithSource(i.installDir, SourceInstalled, false, i.allowedUpstreams, false)
	for _, warning := range warnings {
		i.log.Warn("skipping installed plugin", "err", warning)
	}
	if err != nil {
		return nil, err
	}
	loaded := make([]*Plugin, 0, len(plugins))
	for _, p := range plugins {
		if i.registry.IsBundled(p.Manifest.ID) {
			i.log.Warn("installed plugin id collides with a bundled plugin; skipping", "id", p.Manifest.ID)
			continue
		}
		i.registry.Upsert(p)
		loaded = append(loaded, p)
	}
	if err := i.registry.SyncPermissions(context.Background()); err != nil {
		return nil, err
	}
	return loaded, nil
}

// InstallFromGit clones a public HTTPS repository containing a prebuilt plugin.
func (i *Installer) InstallFromGit(ctx context.Context, rawURL, ref string, replace bool) (*Plugin, error) {
	if err := i.validateGitURL(rawURL); err != nil {
		return nil, err
	}
	staging, err := os.MkdirTemp(i.installDir, ".tmp-*")
	if err != nil {
		return nil, biz.ErrInternal("failed to create staging directory")
	}
	defer os.RemoveAll(staging)

	ref = strings.TrimSpace(ref)
	args := []string{"clone", "--depth", "1", "--no-tags", "--single-branch"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, rawURL, staging)

	cloneCtx, cancel := context.WithTimeout(ctx, gitCloneTimeout)
	defer cancel()
	cmd := exec.CommandContext(cloneCtx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=", "GIT_CONFIG_NOSYSTEM=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, biz.ErrInvalidArgument("git clone failed: " + firstLine(string(output)))
	}
	commit := strings.TrimSpace(i.runGit(ctx, staging, "rev-parse", "HEAD"))
	_ = os.RemoveAll(filepath.Join(staging, ".git"))

	return i.finalize(staging, installMeta{Source: "git", URL: rawURL, Ref: ref, Commit: commit}, replace)
}

// InstallFromArchive installs a prebuilt plugin from a zip archive.
func (i *Installer) InstallFromArchive(ctx context.Context, data []byte, replace bool) (*Plugin, error) {
	if len(data) == 0 {
		return nil, biz.ErrInvalidArgument("archive is empty")
	}
	if len(data) > maxArchiveBytes {
		return nil, biz.ErrInvalidArgument("archive exceeds 10 MiB")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, biz.ErrInvalidArgument("invalid zip archive")
	}
	staging, err := os.MkdirTemp(i.installDir, ".tmp-*")
	if err != nil {
		return nil, biz.ErrInternal("failed to create staging directory")
	}
	defer os.RemoveAll(staging)
	if err := extractZip(reader, staging); err != nil {
		return nil, err
	}
	root, err := pluginRoot(staging)
	if err != nil {
		return nil, biz.ErrInvalidArgument("archive is empty")
	}
	return i.finalize(root, installMeta{Source: "upload"}, replace)
}

// Uninstall removes an installed plugin. Bundled plugins are protected.
func (i *Installer) Uninstall(ctx context.Context, id string) error {
	p, ok := i.registry.Get(id)
	if !ok {
		return biz.ErrPluginNotFound()
	}
	if p.Source != SourceInstalled {
		return biz.ErrForbidden()
	}
	dir, err := filepath.Abs(p.Dir)
	if err != nil {
		return biz.ErrInternal("invalid plugin directory")
	}
	base, err := filepath.Abs(i.installDir)
	if err != nil {
		return biz.ErrInternal("invalid install directory")
	}
	if !strings.HasPrefix(dir, base+string(os.PathSeparator)) {
		return biz.ErrForbidden()
	}
	if err := os.RemoveAll(dir); err != nil {
		return biz.ErrInternal("failed to remove plugin directory")
	}
	i.registry.Remove(id)
	return i.registry.SyncPermissions(ctx)
}

func (i *Installer) finalize(root string, meta installMeta, replace bool) (*Plugin, error) {
	raw, err := os.ReadFile(filepath.Join(root, "plugin.json"))
	if err != nil {
		return nil, biz.ErrInvalidArgument("plugin.json is required")
	}
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, biz.ErrInvalidArgument("invalid plugin.json")
	}
	if err := manifest.Validate(manifest.ID, false, i.allowedUpstreams); err != nil {
		return nil, biz.ErrInvalidArgument("invalid plugin manifest: " + err.Error())
	}
	entry := filepath.Join(root, "dist", filepath.FromSlash(manifest.Frontend.Entry))
	if _, err := os.Stat(entry); err != nil {
		return nil, biz.ErrInvalidArgument("plugin dist/" + manifest.Frontend.Entry + " is required")
	}
	if i.registry.IsBundled(manifest.ID) {
		return nil, biz.ErrForbidden()
	}

	target := filepath.Join(i.installDir, manifest.ID)
	backup := ""
	if _, err := os.Stat(target); err == nil {
		if !replace {
			return nil, biz.ErrInvalidArgument("plugin already installed; set replace=true")
		}
		backup = target + ".bak-" + randomSuffix()
		if err := os.Rename(target, backup); err != nil {
			return nil, biz.ErrInternal("failed to back up the existing plugin")
		}
	}
	if err := os.Rename(root, target); err != nil {
		if backup != "" {
			_ = os.Rename(backup, target)
		}
		return nil, biz.ErrInternal("failed to move the plugin into place")
	}
	if backup != "" {
		_ = os.RemoveAll(backup)
	}

	meta.InstalledAt = time.Now().UTC()
	if encoded, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(target, ".install.json"), encoded, 0o644)
	}

	p := &Plugin{Manifest: &manifest, Dir: target, Source: SourceInstalled}
	i.registry.Upsert(p)
	if err := i.registry.SyncPermissions(context.Background()); err != nil {
		return nil, err
	}
	i.log.Info("installed plugin", "id", manifest.ID, "version", manifest.Version, "source", meta.Source)
	return p, nil
}

func (i *Installer) validateGitURL(raw string) error {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return biz.ErrInvalidArgument("git url must be an https URL")
	}
	if len(i.allowedRepos) == 0 {
		return nil
	}
	withoutScheme := strings.TrimPrefix(raw, "https://")
	for _, allowed := range i.allowedRepos {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}
		allowed = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(allowed, "https://"), "http://"), "/")
		if strings.EqualFold(parsed.Host, allowed) || strings.HasPrefix(withoutScheme, allowed+"/") {
			return nil
		}
	}
	return biz.ErrForbidden()
}

func (i *Installer) runGit(parent context.Context, dir string, args ...string) string {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	out, _ := cmd.Output()
	return string(out)
}

func extractZip(reader *zip.Reader, dest string) error {
	if len(reader.File) > maxArchiveFiles {
		return biz.ErrInvalidArgument("archive has too many files")
	}
	var total int64
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return biz.ErrInvalidArgument("archive contains a symlink")
		}
		name := strings.ReplaceAll(file.Name, "\\", "/")
		clean := path.Clean("/" + name)
		if clean == "/" || clean == "." {
			continue
		}
		relative := strings.TrimPrefix(clean, "/")
		if strings.Contains(relative, ":") {
			return biz.ErrInvalidArgument("archive contains an unsafe path")
		}
		target := filepath.Join(dest, filepath.FromSlash(relative))
		rel, err := filepath.Rel(dest, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return biz.ErrInvalidArgument("archive contains an unsafe path")
		}
		if file.UncompressedSize64 > maxArchiveFile {
			return biz.ErrInvalidArgument("archive file is too large")
		}
		total += int64(file.UncompressedSize64)
		if total > maxExtractBytes {
			return biz.ErrInvalidArgument("archive expands beyond the limit")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return biz.ErrInternal("failed to extract the archive")
		}
		src, err := file.Open()
		if err != nil {
			return biz.ErrInvalidArgument("invalid archive entry")
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			_ = src.Close()
			return biz.ErrInternal("failed to write the archive entry")
		}
		written, copyErr := io.Copy(out, io.LimitReader(src, maxArchiveFile+1))
		_ = out.Close()
		_ = src.Close()
		if copyErr != nil {
			return biz.ErrInternal("failed to extract the archive")
		}
		if written > maxArchiveFile {
			return biz.ErrInvalidArgument("archive file is too large")
		}
	}
	return nil
}

func pluginRoot(staging string) (string, error) {
	entries, err := os.ReadDir(staging)
	if err != nil {
		return "", err
	}
	filtered := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == "__MACOSX" || entry.Name() == ".DS_Store" {
			continue
		}
		filtered = append(filtered, entry)
	}
	if len(filtered) == 1 && filtered[0].IsDir() {
		return filepath.Join(staging, filtered[0].Name()), nil
	}
	if len(filtered) == 0 {
		return "", fmt.Errorf("empty archive")
	}
	return staging, nil
}

func randomSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if index := strings.IndexByte(value, '\n'); index >= 0 {
		return strings.TrimSpace(value[:index])
	}
	if len(value) > 200 {
		return value[:200]
	}
	return value
}
