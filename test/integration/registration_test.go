//go:build integration

package integration

// TestYAMLTypesAllRegistered cross-checks that every *.yaml under types/
// has a matching registration in internal/probes.Register(). This closes
// the loophole where a type YAML ships in the image but no probe is wired
// (the SDK would respond with "unknown type" for every probe of it). Fast
// enough to run without external deps.
//
// Run with:
//
//	go test -tags=integration ./test/integration/ -run TestYAMLTypesAllRegistered -v

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt-provider-aws/internal/probes"
	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestYAMLTypesAllRegistered(t *testing.T) {
	typesDir, err := filepath.Abs("../../types")
	if err != nil {
		t.Fatalf("resolve types dir: %v", err)
	}
	entries, err := os.ReadDir(typesDir)
	if err != nil {
		t.Fatalf("read types dir: %v", err)
	}
	var yamlTypes []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		yamlTypes = append(yamlTypes, strings.TrimSuffix(e.Name(), ".yaml"))
	}
	sort.Strings(yamlTypes)

	r := provider.NewRegistry()
	probes.Register(r)

	for _, typ := range yamlTypes {
		// Probe with a fake fact — we only care whether the type is known.
		_, err := r.Probe(t.Context(), provider.Request{
			Type: typ,
			Name: "n",
			Fact: "__does_not_exist__",
		})
		if err == nil {
			continue
		}
		if strings.Contains(err.Error(), "unknown type") {
			t.Errorf("type %q has a YAML file but no probe registration in internal/probes.Register", typ)
		}
		// Any other error (unknown fact, usage, etc.) means the type IS
		// registered — the probe just rejected the fake fact name.
	}
}
