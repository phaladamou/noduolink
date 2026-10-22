package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config représente la structure principale du fichier de configuration de NoduoLink.
type Config struct {
	// LogLevel définit le niveau de log (debug, info, warn, error).
	LogLevel string `yaml:"log_level"`
	
	// Coordinator contient les paramètres pour se connecter au serveur central.
	Coordinator CoordinatorConfig `yaml:"coordinator"`
	
	// Identity contient les paramètres liés à l'identité de l'appareil.
	Identity IdentityConfig `yaml:"identity"`
}

// CoordinatorConfig définit comment l'agent rejoint le coordinateur.
type CoordinatorConfig struct {
	URL string `yaml:"url"`
}

// IdentityConfig définit les chemins locaux pour les clés.
type IdentityConfig struct {
	PrivateKeyPath string `yaml:"private_key_path"`
}

// Load lit le fichier de configuration depuis le chemin spécifié.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le fichier de config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("erreur de parsing YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration invalide: %w", err)
	}

	return &cfg, nil
}

// Validate vérifie que les champs obligatoires sont présents et valides.
func (c *Config) Validate() error {
	if c.LogLevel == "" {
		c.LogLevel = "info" // Valeur par défaut
	}
	
	if c.Coordinator.URL == "" {
		return fmt.Errorf("coordinator.url est obligatoire")
	}
	
	if c.Identity.PrivateKeyPath == "" {
		return fmt.Errorf("identity.private_key_path est obligatoire")
	}

	return nil
}
