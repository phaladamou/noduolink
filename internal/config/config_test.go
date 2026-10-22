package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	validYAML := `
log_level: debug
coordinator:
  url: "wss://localhost:8080"
identity:
  private_key_path: "/tmp/key"
`
	os.WriteFile(configPath, []byte(validYAML), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() a échoué: %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel attendu debug, obtenu %s", cfg.LogLevel)
	}
}

func TestLoad_UnknownField(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// 'coordinater' est une faute de frappe volontaire
	invalidYAML := `
coordinater:
  url: "wss://localhost:8080"
identity:
  private_key_path: "/tmp/key"
`
	os.WriteFile(configPath, []byte(invalidYAML), 0644)

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() aurait dû échouer à cause du champ inconnu 'coordinater'")
	}
	if !strings.Contains(err.Error(), "champ inconnu") && !strings.Contains(err.Error(), "field") {
		t.Errorf("L'erreur ne mentionne pas le champ inconnu: %v", err)
	}
}

func TestValidate_BadScheme(t *testing.T) {
	cfg := &Config{
		Coordinator: CoordinatorConfig{URL: "http://insecure.com"}, // http au lieu de https/wss
		Identity:    IdentityConfig{PrivateKeyPath: "/tmp/key"},
	}
	cfg.ApplyDefaults()

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() aurait dû échouer avec une URL http://")
	}
}
