package plugin

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	commonv1 "github.com/example/scaffold/services/api/common/v1"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestInstaller(t *testing.T) (*Installer, *Registry, string) {
	t.Helper()
	dir := t.TempDir()
	registry := &Registry{plugins: map[string]*Plugin{}, log: testLogger()}
	installer := &Installer{installDir: dir, registry: registry, log: testLogger()}
	return installer, registry, dir
}

func buildZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

const demoManifest = `{
  "id": "demo-plugin",
  "name": "Demo",
  "version": "1.0.0",
  "frontend": { "entry": "index.js" }
}`

func reasonOf(err error) string {
	return kerrors.FromError(err).Reason
}

func TestInstallArchiveAndUninstall(t *testing.T) {
	installer, registry, dir := newTestInstaller(t)
	archive := buildZip(t, map[string]string{
		"plugin.json":   demoManifest,
		"dist/index.js": "export default {}",
	})

	if _, err := installer.InstallFromArchive(context.Background(), archive, false); err != nil {
		t.Fatalf("install: %v", err)
	}
	plugin, ok := registry.Get("demo-plugin")
	if !ok || plugin.Source != SourceInstalled {
		t.Fatalf("plugin not registered: %+v", plugin)
	}
	if _, err := os.Stat(filepath.Join(dir, "demo-plugin", ".install.json")); err != nil {
		t.Fatalf("install metadata missing: %v", err)
	}

	if _, err := installer.InstallFromArchive(context.Background(), archive, false); reasonOf(err) != commonv1.ErrorReason_INVALID_ARGUMENT.String() {
		t.Fatalf("duplicate install reason = %q", reasonOf(err))
	}
	if _, err := installer.InstallFromArchive(context.Background(), archive, true); err != nil {
		t.Fatalf("replace install: %v", err)
	}

	if err := installer.Uninstall(context.Background(), "demo-plugin"); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if _, ok := registry.Get("demo-plugin"); ok {
		t.Fatal("plugin still registered after uninstall")
	}
	if _, err := os.Stat(filepath.Join(dir, "demo-plugin")); !os.IsNotExist(err) {
		t.Fatalf("plugin directory still exists: %v", err)
	}
}

func TestInstallRequiresManifestAndDist(t *testing.T) {
	installer, _, _ := newTestInstaller(t)

	noManifest := buildZip(t, map[string]string{"dist/index.js": "x"})
	if _, err := installer.InstallFromArchive(context.Background(), noManifest, false); err == nil {
		t.Fatal("archive without plugin.json was accepted")
	}

	noDist := buildZip(t, map[string]string{"plugin.json": demoManifest})
	if _, err := installer.InstallFromArchive(context.Background(), noDist, false); err == nil {
		t.Fatal("archive without dist/ was accepted")
	}
}

func TestInstallRejectsUnsafePaths(t *testing.T) {
	installer, _, dir := newTestInstaller(t)
	archive := buildZip(t, map[string]string{
		"plugin.json":   demoManifest,
		"dist/index.js": "x",
		"../escape.txt": "escape",
	})
	if _, err := installer.InstallFromArchive(context.Background(), archive, false); err != nil {
		t.Fatalf("install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "..", "escape.txt")); !os.IsNotExist(err) {
		t.Fatal("archive escaped the install directory")
	}

	drivePath := buildZip(t, map[string]string{
		"plugin.json":    demoManifest,
		"dist/index.js":  "x",
		"C:/evil.txt":    "evil",
	})
	if _, err := installer.InstallFromArchive(context.Background(), drivePath, false); err == nil {
		t.Fatal("archive with a drive path was accepted")
	}
}

func TestUninstallRejectsBundledPlugin(t *testing.T) {
	installer, registry, _ := newTestInstaller(t)
	registry.Upsert(&Plugin{
		Manifest: &Manifest{ID: "bundled-plugin"},
		Dir:      filepath.Join("..", "runtime-plugins", "bundled-plugin"),
		Source:   SourceBundled,
	})
	if err := installer.Uninstall(context.Background(), "bundled-plugin"); reasonOf(err) != commonv1.ErrorReason_FORBIDDEN.String() {
		t.Fatalf("reason = %q", reasonOf(err))
	}
}

func TestValidateGitURL(t *testing.T) {
	installer := &Installer{allowedRepos: nil, log: testLogger()}
	if err := installer.validateGitURL("http://github.com/acme/plugin"); err == nil {
		t.Fatal("http url was accepted")
	}
	if err := installer.validateGitURL("https://github.com/acme/plugin"); err != nil {
		t.Fatalf("https url rejected: %v", err)
	}

	restricted := &Installer{allowedRepos: []string{"github.com"}, log: testLogger()}
	if err := restricted.validateGitURL("https://github.com/acme/plugin"); err != nil {
		t.Fatalf("allowlisted url rejected: %v", err)
	}
	if err := restricted.validateGitURL("https://gitlab.com/acme/plugin"); reasonOf(err) != commonv1.ErrorReason_FORBIDDEN.String() {
		t.Fatalf("non-allowlisted reason = %q", reasonOf(err))
	}
}
