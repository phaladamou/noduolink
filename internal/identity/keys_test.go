package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() a échoué: %v", err)
	}

	// Vérifier la taille des clés (Curve25519 = 32 octets)
	if len(kp.PrivateKey) != 32 {
		t.Errorf("La clé privée doit faire 32 octets, obtenu %d", len(kp.PrivateKey))
	}
	if len(kp.PublicKey) != 32 {
		t.Errorf("La clé publique doit faire 32 octets, obtenu %d", len(kp.PublicKey))
	}

	// Vérifier que deux clés générées ne sont pas identiques
	kp2, _ := GenerateKeyPair()
	if string(kp.PrivateKey) == string(kp2.PrivateKey) {
		t.Error("Deux clés privées générées sont identiques (grave problème d'aléatoire)")
	}
}

func TestSaveToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "keys", "device.key")

	kp, _ := GenerateKeyPair()
	err := kp.SaveToDisk(keyPath)
	if err != nil {
		t.Fatalf("SaveToDisk() a échoué: %v", err)
	}

	// Vérifier que le fichier existe
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Le fichier de clé n'existe pas: %v", err)
	}

	// Vérifier les permissions (0600 = lecture/écriture propriétaire uniquement)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Permissions du fichier incorrectes: attendu 0600, obtenu %o", info.Mode().Perm())
	}
}
