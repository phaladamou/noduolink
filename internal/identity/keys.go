package identity

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeyPair représente une paire de clés cryptographiques.
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// GenerateKeyPair génère une nouvelle paire de clés Curve25519.
func GenerateKeyPair() (*KeyPair, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("impossible de générer la clé privée: %w", err)
	}
	return &KeyPair{
		PrivateKey: priv.Bytes(),
		PublicKey:  priv.PublicKey().Bytes(),
	}, nil
}

// LoadFromDisk lit un fichier de clés existant et reconstruit la paire.
func LoadFromDisk(path string) (*KeyPair, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le fichier de clés: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var privBase64, pubBase64 string

	for _, line := range lines {
		if strings.HasPrefix(line, "PRIVATE_KEY=") {
			privBase64 = strings.TrimPrefix(line, "PRIVATE_KEY=")
		} else if strings.HasPrefix(line, "PUBLIC_KEY=") {
			pubBase64 = strings.TrimPrefix(line, "PUBLIC_KEY=")
		}
	}

	if privBase64 == "" || pubBase64 == "" {
		return nil, fmt.Errorf("fichier de clés invalide ou incomplet")
	}

	privBytes, err := base64.StdEncoding.DecodeString(privBase64)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la clé privée: %w", err)
	}

	pubBytes, err := base64.StdEncoding.DecodeString(pubBase64)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la clé publique: %w", err)
	}

	if len(privBytes) != 32 || len(pubBytes) != 32 {
		return nil, fmt.Errorf("taille de clé invalide")
	}

	return &KeyPair{
		PrivateKey: privBytes,
		PublicKey:  pubBytes,
	}, nil
}

// SaveToDisk sauvegarde la paire de clés sur le disque avec des permissions restrictives.
func (kp *KeyPair) SaveToDisk(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("impossible de créer le dossier %s: %w", dir, err)
	}
	privBase64 := base64.StdEncoding.EncodeToString(kp.PrivateKey)
	pubBase64 := base64.StdEncoding.EncodeToString(kp.PublicKey)
	content := fmt.Sprintf("PRIVATE_KEY=%s\nPUBLIC_KEY=%s\n", privBase64, pubBase64)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("impossible d'écrire la clé sur le disque: %w", err)
	}
	return nil
}

// GenerateNodeID génère un identifiant basé sur la clé publique (le même que le serveur).
func GenerateNodeID(publicKeyBase64 string) string {
	hash := sha256.Sum256([]byte(publicKeyBase64))
	return hex.EncodeToString(hash[:])
}

// GenerateLocalNodeID (alias pour la compatibilité).
func GenerateLocalNodeID(publicKeyBase64 string) string {
	return GenerateNodeID(publicKeyBase64)
}
