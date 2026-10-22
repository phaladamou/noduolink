package identity

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// KeyPair représente une paire de clés cryptographiques (publique et privée).
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// GenerateKeyPair génère une nouvelle paire de clés Curve25519.
// Curve25519 est le standard utilisé par WireGuard pour sa sécurité.
func GenerateKeyPair() (*KeyPair, error) {
	// Générer la clé privée
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("impossible de générer la clé privée: %w", err)
	}

	return &KeyPair{
		PrivateKey: priv.Bytes(),
		PublicKey:  priv.PublicKey().Bytes(),
	}, nil
}

// SaveToDisk sauvegarde la clé privée sur le disque local avec des permissions restrictives.
// La clé publique n'a pas besoin d'être gardée secrète, mais on la sauvegarde à côté.
func (kp *KeyPair) SaveToDisk(path string) error {
	// Créer le dossier parent si nécessaire
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("impossible de créer le dossier %s: %w", dir, err)
	}

	// Encoder en base64 pour que ce soit lisible (comme le fait WireGuard)
	privBase64 := base64.StdEncoding.EncodeToString(kp.PrivateKey)
	pubBase64 := base64.StdEncoding.EncodeToString(kp.PublicKey)

	// Format simple : clé privée, puis clé publique
	content := fmt.Sprintf("PRIVATE_KEY=%s\nPUBLIC_KEY=%s\n", privBase64, pubBase64)

	// Écrire le fichier avec permissions 0600 (seul le propriétaire peut lire/écrire)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("impossible d'écrire la clé sur le disque: %w", err)
	}

	return nil
}
