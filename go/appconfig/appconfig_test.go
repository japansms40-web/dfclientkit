package appconfig

import (
	"path/filepath"
	"testing"
)

type testConfig struct {
	Threads int    `json:"threads"`
	Name    string `json:"name"`
}

func TestLoadFileSaveFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	cfg := testConfig{Threads: 5, Name: "hi"}
	if err := SaveFile(path, cfg); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}
	got := LoadFile(path, testConfig{})
	if got != cfg {
		t.Errorf("LoadFile() = %+v, want %+v", got, cfg)
	}
}

func TestLoadFileMissingReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	defaults := testConfig{Threads: 1, Name: "default"}
	if got := LoadFile(path, defaults); got != defaults {
		t.Errorf("LoadFile() = %+v, want %+v", got, defaults)
	}
}

func TestLoadFileCorruptJSONReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := SaveFile(path, "not-json-shaped-but-valid-string"); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}
	defaults := testConfig{Threads: 9, Name: "fallback"}
	got := LoadFile(path, defaults)
	if got != defaults {
		t.Errorf("LoadFile() = %+v, want %+v（JSON 字符串不是合法的 testConfig 形状应回退默认值）", got, defaults)
	}
}

func TestDefaultDirIncludesAppName(t *testing.T) {
	dir, err := DefaultDir("ghpublisher")
	if err != nil {
		t.Fatalf("DefaultDir() error = %v", err)
	}
	if filepath.Base(dir) != "ghpublisher" {
		t.Errorf("DefaultDir() = %q, want basename ghpublisher", dir)
	}
}
