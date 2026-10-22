package identity

import (
	"path/filepath"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() a échoué: %v", err)
	}
	if len(kp.PrivateKey) != 32 || len(kp.PublicKey) != 32 {
		t.Error("La taille des clés est incorrecte")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "device.key")

	kp1, _ := GenerateKeyPair()
	if err := kp1.SaveToDisk(keyPath); err != nil {
		t.Fatalf("SaveToDisk() a échoué: %v", err)
	}

	kp2, err := LoadFromDisk(keyPath)
	if err != nil {
		t.Fatalf("LoadFromDisk() a échoué: %v", err)
	}

	if string(kp1.PrivateKey) != string(kp2.PrivateKey) {
		t.Error("Les clés privées ne correspondent pas après rechargement")
	}
}
