package rbac

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

type permissionEntry struct {
	Code     string `yaml:"code"`
	Name     string `yaml:"name"`
	Resource string `yaml:"resource"`
	Action   string `yaml:"action"`
}

type permissionSpec struct {
	Permissions []permissionEntry `yaml:"permissions"`
}

func TestCatalogMatchesYAML(t *testing.T) {
	raw, err := os.ReadFile("../../../../contracts/permissions.yaml")
	if err != nil {
		t.Fatalf("read permission catalog: %v", err)
	}
	var spec permissionSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parse permission catalog: %v", err)
	}
	if len(spec.Permissions) == 0 {
		t.Fatal("permission catalog is empty")
	}
	if len(spec.Permissions) != len(Catalog()) {
		t.Fatalf("catalog has %d entries, yaml has %d", len(Catalog()), len(spec.Permissions))
	}
	want := make(map[string]permissionEntry, len(spec.Permissions))
	for _, entry := range spec.Permissions {
		want[entry.Code] = entry
	}
	for _, def := range Catalog() {
		entry, ok := want[def.Code]
		if !ok {
			t.Fatalf("generated permission %q is not in the yaml", def.Code)
		}
		if entry.Name != def.Name || entry.Resource != def.Resource || entry.Action != def.Action {
			t.Fatalf("permission %q metadata drifted: yaml=%+v go=%+v", def.Code, entry, def)
		}
	}
}
