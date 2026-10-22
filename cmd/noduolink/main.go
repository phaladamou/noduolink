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
	flag.Parse()

	log.SetPrefix("[noduolink] ")
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
		// Note: Pour l'instant, on ne relit pas les clés sur le disque, on en génère une nouvelle pour le test
		// On corrigera ça dans la prochaine étape.
		kp, _ = identity.GenerateKeyPair()
	}

	pubKeyBase64 := base64.StdEncoding.EncodeToString(kp.PublicKey)

	fmt.Println("--------------------------------------------------")
	fmt.Println("  NoduoLink Client - Your network. Everywhere.")
	fmt.Println("--------------------------------------------------")

	// 2. Connexion au Coordinateur (WebSocket)
	// On utilise ws:// pour le test local, même si la config dit wss://
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
		NodeID:    "device-test-01",
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

	fmt.Println("--------------------------------------------------")
	fmt.Println("  Statut : Connecté au Coordinateur")
	fmt.Println("--------------------------------------------------")
}
