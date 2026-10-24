package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Load scans a plugins directory and returns every valid manifest. Invalid
// plugins are skipped when strict is false.
func Load(root string, dev bool, allowedUpstreams []string, strict bool) ([]*Plugin, []error, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read plugins directory: %w", err)
	}

	plugins := make([]*Plugin, 0, len(entries))
	var warnings []error
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		manifestPath := filepath.Join(dir, "plugin.json")
		raw, err := os.ReadFile(manifestPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			warnings = append(warnings, fmt.Errorf("read %s: %w", manifestPath, err))
			continue
		}
		var manifest Manifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			warnings = append(warnings, fmt.Errorf("parse %s: %w", manifestPath, err))
			continue
		}
		if err := manifest.Validate(entry.Name(), dev, allowedUpstreams); err != nil {
			warnings = append(warnings, fmt.Errorf("validate %s: %w", manifestPath, err))
			continue
		}
		plugins = append(plugins, &Plugin{Manifest: &manifest, Dir: dir})
	}

	if strict && len(warnings) > 0 {
		return nil, warnings, fmt.Errorf("invalid plugins: %v", warnings)
	}
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].Manifest.ID < plugins[j].Manifest.ID })
	return plugins, warnings, nil
}
