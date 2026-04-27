package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfigIncludesCodexProvider(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	cfg := DefaultConfig()
	codex, ok := cfg.Providers["codex"]
	if !ok {
		t.Fatalf("DefaultConfig().Providers missing codex")
	}

	if codex.Enabled {
		t.Fatalf("codex Enabled = true, want false")
	}

	want := filepath.Join("/home/alice", ".codex", "skills")
	if codex.Path != want {
		t.Fatalf("codex Path = %q, want %q", codex.Path, want)
	}
}

func TestDefaultConfigProviderDefaults(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	cfg := DefaultConfig()
	tests := map[string]bool{
		"claude":   true,
		"cursor":   true,
		"qoder":    true,
		"windsurf": false,
		"copilot":  false,
		"opencode": false,
		"codex":    false,
	}

	for name, wantEnabled := range tests {
		provider, ok := cfg.Providers[name]
		if !ok {
			t.Fatalf("provider %q missing from DefaultConfig", name)
		}
		if provider.Enabled != wantEnabled {
			t.Fatalf("provider %q Enabled = %v, want %v", name, provider.Enabled, wantEnabled)
		}
	}
}
