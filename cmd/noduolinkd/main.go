package main

import (
	"flag"
	"log"
	"os"

	"github.com/phaladamou/noduolink/internal/config"
)

func main() {
	// Définir le flag pour le fichier de configuration
	configPath := flag.String("config", "config.yaml", "Chemin vers le fichier de configuration")
	flag.Parse()

	log.SetPrefix("[noduolinkd] ")
	log.Println("Démarrage du Coordinateur NoduoLink...")

	// 1. Charger la configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Erreur critique de configuration: %v", err)
	}
	log.Printf("Configuration chargée avec succès (LogLevel: %s)", cfg.LogLevel)

	// 2. Vérifier l'URL du coordinateur (pour s'assurer que le serveur sait où il écoute)
	log.Printf("Le serveur est configuré pour écouter sur: %s", cfg.Coordinator.URL)

	// 3. Simuler le démarrage du serveur (on le fera à l'étape suivante)
	log.Println("Coordinateur initialisé. En attente de connexions (simulation)...")

	// Pour l'instant, on quitte proprement
	log.Println("Arrêt du Coordinateur.")
	os.Exit(0)
}
