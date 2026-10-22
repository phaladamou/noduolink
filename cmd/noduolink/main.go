package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/phaladamou/noduolink/internal/config"
	"github.com/phaladamou/noduolink/internal/identity"
)

func main() {
	// Définir les flags de la CLI
	configPath := flag.String("config", "config.yaml", "Chemin vers le fichier de configuration")
	flag.Parse()

	log.SetPrefix("[noduolink] ")
	log.Println("Démarrage du Client NoduoLink...")

	// 1. Charger la configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Erreur critique de configuration: %v", err)
	}

	// 2. Gérer l'identité (Vérifier si les clés existent, sinon les créer)
	keyPath := cfg.Identity.PrivateKeyPath
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Printf("Aucune identité trouvée à %s. Génération d'une nouvelle paire de clés...", keyPath)

		kp, err := identity.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Impossible de générer les clés: %v", err)
		}

		if err := kp.SaveToDisk(keyPath); err != nil {
			log.Fatalf("Impossible de sauvegarder les clés: %v", err)
		}
		log.Println("Nouvelle identité cryptographique générée et sécurisée.")
	} else {
		log.Printf("Identité existante chargée depuis: %s", keyPath)
	}

	// 3. Afficher le statut
	fmt.Println("--------------------------------------------------")
	fmt.Println("  NoduoLink Client - Your network. Everywhere.")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Statut       : Prêt\n")
	fmt.Printf("  Coordinateur : %s\n", cfg.Coordinator.URL)
	fmt.Printf("  Clé locale   : %s\n", keyPath)
	fmt.Println("--------------------------------------------------")
	log.Println("Client initialisé avec succès.")
}
