package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	// Test 1 : Configuration vide (doit échouer)
	cfg := &Config{}
	if err := cfg.Validate(); err == nil {
		t.Errorf("Validate() aurait dû échouer avec une config vide")
	}

	// Test 2 : Configuration valide
	cfg = &Config{
		Coordinator: CoordinatorConfig{URL: "wss://coordinator.noduolink.com"},
		Identity:    IdentityConfig{PrivateKeyPath: "/etc/noduolink/keys/private.key"},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() a échoué avec une config valide: %v", err)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel par défaut non appliqué")
	}
}

func TestLoad(t *testing.T) {
	// Créer un fichier temporaire pour le test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	validYAML := `
log_level: debug
coordinator:
  url: "wss://localhost:8080"
identity:
  private_key_path: "/tmp/key"
`
	if err := os.WriteFile(configPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Impossible de créer le fichier de test: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() a échoué: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel attendu debug, obtenu %s", cfg.LogLevel)
	}
	if cfg.Coordinator.URL != "wss://localhost:8080" {
		t.Errorf("URL incorrecte")
	}
}
