package provider

import (
	"path/filepath"
	"testing"
)

func TestDefinitionsIncludeCodex(t *testing.T) {
	defs := Definitions()

	var found bool
	for _, def := range defs {
		if def.Name == "codex" {
			found = true
			if def.DefaultEnabled {
				t.Fatalf("codex DefaultEnabled = true, want false")
			}
			got := def.Path("/home/alice")
			want := filepath.Join("/home/alice", ".codex", "skills")
			if got != want {
				t.Fatalf("codex path = %q, want %q", got, want)
			}
		}
	}

	if !found {
		t.Fatalf("Definitions() did not include codex")
	}
}

func TestDefinitionsReturnCopy(t *testing.T) {
	defs := Definitions()
	defs[0].Name = "changed"

	fresh := Definitions()
	if fresh[0].Name == "changed" {
		t.Fatalf("Definitions() returned mutable shared backing array")
	}
}
