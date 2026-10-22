package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/phaladamou/noduolink/api/proto"
	"github.com/phaladamou/noduolink/internal/config"
	"github.com/phaladamou/noduolink/internal/identity"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Chemin vers le fichier de configuration")
	nodeID := flag.String("id", "device-test-01", "ID de l'appareil (pour simuler plusieurs machines)")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[noduolink-%s] ", *nodeID))
	log.Println("Démarrage du Client NoduoLink...")

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Erreur critique de configuration: %v", err)
	}

	// 1. Gestion de l'identité
	keyPath := cfg.Identity.PrivateKeyPath
	var kp *identity.KeyPair

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Printf("Génération d'une nouvelle identité...")
		kp, err = identity.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Impossible de générer les clés: %v", err)
		}
		if err := kp.SaveToDisk(keyPath); err != nil {
			log.Fatalf("Impossible de sauvegarder les clés: %v", err)
		}
	} else {
		log.Printf("Identité chargée depuis: %s", keyPath)
		kp, _ = identity.GenerateKeyPair() // Note: on régénère pour l'instant
	}

	pubKeyBase64 := base64.StdEncoding.EncodeToString(kp.PublicKey)

	fmt.Println("--------------------------------------------------")
	fmt.Printf("  NoduoLink Client - %s\n", *nodeID)
	fmt.Println("--------------------------------------------------")

	// 2. Connexion au Coordinateur
	serverURL := "ws://localhost:8080/ws"
	log.Printf("Tentative de connexion au coordinateur sur %s...", serverURL)

	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Impossible de se connecter au coordinateur: %v", err)
	}
	defer conn.Close()
	log.Println("Connecté au coordinateur !")

	// 3. Envoyer le message REGISTER
	regReq := proto.RegisterRequest{
		NodeID:    *nodeID,
		PublicKey: pubKeyBase64,
	}
	msg := proto.Message{
		Type: proto.TypeRegister,
		Data: regReq,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Fatalf("Erreur lors de l'envoi du REGISTER: %v", err)
	}
	log.Println("Message REGISTER envoyé. En attente de l'ACK...")

	// 4. Attendre la réponse (ACK)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var ack proto.Message
	if err := conn.ReadJSON(&ack); err != nil {
		log.Fatalf("Erreur lors de la lecture de l'ACK: %v", err)
	}

	if ack.Type == proto.TypeAck {
		log.Printf("SUCCÈS: Réponse reçue du serveur: %v", ack.Data)
	} else {
		log.Printf("Réponse inattendue: %s", ack.Type)
	}

	// 5. Désactiver la limite de temps pour garder la connexion ouverte
	conn.SetReadDeadline(time.Time{})

	fmt.Println("--------------------------------------------------")
	fmt.Println("  Statut : Connecté au Coordinateur (en attente...)")
	fmt.Println("--------------------------------------------------")

	// 6. Boucle de lecture (garder la connexion active)
	for {
		var incoming proto.Message
		if err := conn.ReadJSON(&incoming); err != nil {
			log.Printf("Déconnexion du serveur: %v", err)
			break
		}
		log.Printf("Message reçu: %s", incoming.Type)
	}
}
