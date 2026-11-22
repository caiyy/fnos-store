package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFailedSavePreservesLiveConfig(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	old := Config{AccessTokenHash: "original"}
	if err := m.SaveConfig(old); err != nil {
		t.Fatal(err)
	}
	// A directory in place of the file forces the atomic rename to fail,
	// even when tests run as root.
	if err := os.Remove(m.filePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(m.filePath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := m.SaveConfig(Config{AccessTokenHash: "new"}); err == nil {
		t.Fatal("save should fail")
	}
	if m.Get().AccessTokenHash != "original" {
		t.Fatal("failed save changed live token")
	}
	files, err := filepath.Glob(filepath.Join(dir, ".config-*"))
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary files remain: %v %v", files, err)
	}
}

func TestConfigFileIsPrivate(t *testing.T) {
	m := NewManager(t.TempDir())
	if err := m.SaveConfig(Config{}); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(m.filePath)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("permissions: %v", st.Mode())
	}
}

func TestAutoUpdateDefaultsEnabledAndCanBeDisabled(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	if _, err := m.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if !m.Get().AutoUpdate {
		t.Fatal("auto update should default to enabled")
	}

	cfg := m.Get()
	cfg.AutoUpdate = false
	if err := m.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	reloaded := NewManager(dir)
	if _, err := reloaded.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if reloaded.Get().AutoUpdate {
		t.Fatal("explicitly disabled auto update should remain disabled")
	}
}

func TestLegacyHourlyIntervalMigratesToTenMinutes(t *testing.T) {
	dir := t.TempDir()
	legacy := []byte(`{"check_interval_hours":24,"mirror":"gh-proxy","docker_mirror":"daocloud"}`)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), legacy, 0o600); err != nil {
		t.Fatal(err)
	}

	m := NewManager(dir)
	if _, err := m.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if got := m.Get().CheckIntervalMinutes; got != DefaultCheckIntervalMinutes {
		t.Fatalf("legacy interval migrated to %d minutes, want %d", got, DefaultCheckIntervalMinutes)
	}

	if err := m.SaveConfig(m.Get()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte(`"check_interval_minutes": 10`); !bytes.Contains(raw, want) {
		t.Fatalf("saved config does not contain %q: %s", want, raw)
	}
	if bytes.Contains(raw, []byte(`"check_interval_hours"`)) {
		t.Fatalf("saved config still contains legacy interval: %s", raw)
	}
}
