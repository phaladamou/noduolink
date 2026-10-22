package config

import (
	"bytes"
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

// Config représente la structure principale du fichier de configuration de NoduoLink.
type Config struct {
	LogLevel    string            `yaml:"log_level"`
	Coordinator CoordinatorConfig `yaml:"coordinator"`
	Identity    IdentityConfig    `yaml:"identity"`
}

type CoordinatorConfig struct {
	URL string `yaml:"url"`
}

type IdentityConfig struct {
	PrivateKeyPath string `yaml:"private_key_path"`
}

// Load lit, parse et valide le fichier de configuration.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le fichier de config %s: %w", path, err)
	}

	var cfg Config
	// Décodage strict : rejette les champs inconnus
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("erreur de parsing YAML (champ inconnu ou invalide): %w", err)
	}

	// Appliquer les valeurs par défaut AVANT la validation
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration invalide: %w", err)
	}

	return &cfg, nil
}

// ApplyDefaults sépare l'application des valeurs par défaut de la validation.
func (c *Config) ApplyDefaults() {
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
}

// Validate vérifie strictement les champs.
func (c *Config) Validate() error {
	if c.Coordinator.URL == "" {
		return fmt.Errorf("coordinator.url est obligatoire")
	}

	// Validation stricte de l'URL
	u, err := url.Parse(c.Coordinator.URL)
	if err != nil {
		return fmt.Errorf("coordinator.url est mal formée: %w", err)
	}
	if u.Scheme != "wss" && u.Scheme != "https" {
		return fmt.Errorf("coordinator.url doit utiliser le protocole wss:// ou https:// (actuellement: %s)", u.Scheme)
	}

	if c.Identity.PrivateKeyPath == "" {
		return fmt.Errorf("identity.private_key_path est obligatoire")
	}

	return nil
}
