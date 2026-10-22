package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"github.com/phaladamou/noduolink/api/proto"
	"github.com/phaladamou/noduolink/internal/config"
	"github.com/phaladamou/noduolink/internal/identity"
	"github.com/phaladamou/noduolink/internal/tunnel"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Chemin vers le fichier de configuration")
	nodeIDFlag := flag.String("id", "device-test-01", "ID de l'appareil (pour l'affichage)")
	connectTo := flag.String("connect-to", "", "NodeID de la passerelle à laquelle se connecter")
	acceptConnections := flag.Bool("accept", true, "Accepter automatiquement les demandes de connexion")

	// Flags pour WireGuard
	wgPrivateKey := flag.String("wg-private", "", "Clé privée WireGuard (base64)")
	wgPeerPublicKey := flag.String("wg-peer", "", "Clé publique WireGuard du pair (base64)")
	wgLocalIP := flag.String("wg-ip", "10.0.0.2/24", "Adresse IP locale du tunnel")
	wgPeerIP := flag.String("wg-peer-ip", "10.0.0.1/32", "Adresse IP du pair dans le tunnel")
	wgEndpoint := flag.String("wg-endpoint", "", "Adresse IP:port du pair (optionnel)")

	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[noduolink-%s] ", *nodeIDFlag))
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
		kp, err = identity.LoadFromDisk(keyPath)
		if err != nil {
			log.Fatalf("Impossible de charger les clés: %v", err)
		}
		log.Println("Identité chargée depuis le disque.")
	}

	pubKeyBase64 := base64.StdEncoding.EncodeToString(kp.PublicKey)
	localNodeID := identity.GenerateLocalNodeID(pubKeyBase64)

	fmt.Println("--------------------------------------------------")
	fmt.Printf("  NoduoLink Client - %s\n", *nodeIDFlag)
	fmt.Println("--------------------------------------------------")

	// 2. Connexion au Coordinateur
	serverURL := "ws://localhost:8080/ws"
	log.Printf("Tentative de connexion au coordinateur...")

	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Impossible de se connecter: %v", err)
	}
	defer conn.Close()

	// 3. Envoyer REGISTER
	regReq := proto.RegisterRequest{NodeID: *nodeIDFlag, PublicKey: pubKeyBase64}
	msg := proto.Message{Type: proto.TypeRegister, Data: regReq}
	if err := conn.WriteJSON(msg); err != nil {
		log.Fatalf("Erreur d'envoi REGISTER: %v", err)
	}

	// 4. Attendre ACK
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var ack proto.Message
	if err := conn.ReadJSON(&ack); err != nil {
		log.Fatalf("Erreur de lecture ACK: %v", err)
	}
	conn.SetReadDeadline(time.Time{})

	log.Printf("Enregistré avec succès. NodeID: %s", localNodeID[:8])

	// 5. Envoyer la demande de connexion si spécifiée
	if *connectTo != "" {
		log.Printf("Envoi de la demande de connexion vers %s...", *connectTo)
		req := proto.Message{Type: proto.TypeConnectRequest, Data: proto.ConnectRequest{
			TargetNodeID: *connectTo,
		}}
		conn.WriteJSON(req)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("  Statut : Connecté (en attente...)")
	fmt.Println("--------------------------------------------------")

	// Fonction utilitaire pour configurer l'interface réseau locale
	setupLocalInterface := func(ip string) {
		// Ajouter l'adresse IP
		cmd := exec.Command("sudo", "ip", "addr", "add", ip, "dev", "wg0")
		if err := cmd.Run(); err != nil {
			log.Printf("Note: impossible d'ajouter l'IP (peut-être déjà présente): %v", err)
		}
		// Activer l'interface
		exec.Command("sudo", "ip", "link", "set", "wg0", "up").Run()
		log.Printf("Interface wg0 configurée avec l'IP %s", ip)
	}

	// 6. Boucle de lecture
	for {
		var incoming proto.Message
		if err := conn.ReadJSON(&incoming); err != nil {
			log.Printf("Déconnexion du serveur: %v", err)
			break
		}

		switch incoming.Type {
		case proto.TypeConnectOffer:
			log.Println("--- DEMANDE DE CONNEXION ENTRANTE ---")
			dataMap := incoming.Data.(map[string]interface{})
			initiatorID := dataMap["initiator_node_id"].(string)
			log.Printf("Le nœud %s veut se connecter.", initiatorID)

			if *acceptConnections {
				accept := proto.Message{Type: proto.TypeConnectAccept, Data: proto.ConnectResponse{
					TargetNodeID: initiatorID,
					Accepted:     true,
					Message:      "Connexion acceptée",
				}}
				conn.WriteJSON(accept)
				log.Println("Connexion acceptée.")

				// Configurer le tunnel WireGuard en tant que passerelle
				if *wgPrivateKey != "" && *wgPeerPublicKey != "" {
					log.Println("Configuration du tunnel WireGuard (mode Passerelle)...")
					tunnelCfg := tunnel.Config{
						InterfaceName: "wg0",
						PrivateKey:    *wgPrivateKey,
						ListenPort:    51820,
						PeerPublicKey: *wgPeerPublicKey,
						PeerEndpoint:  *wgEndpoint,
						AllowedIPs:    []string{*wgPeerIP},
					}
					if err := tunnel.SetupTunnel(tunnelCfg); err != nil {
						log.Printf("Erreur de configuration du tunnel: %v", err)
					} else {
						setupLocalInterface(*wgLocalIP)
					}
				}
			} else {
				reject := proto.Message{Type: proto.TypeConnectReject, Data: proto.ConnectResponse{
					TargetNodeID: initiatorID,
					Accepted:     false,
					Message:      "Refusée par l'utilisateur",
				}}
				conn.WriteJSON(reject)
				log.Println("Connexion refusée.")
			}

		case proto.TypeConnectAccept:
			log.Println("--- CONNEXION ACCEPTÉE PAR LE PAIR ---")

			// Configurer le tunnel WireGuard en tant que client
			if *wgPrivateKey != "" && *wgPeerPublicKey != "" {
				log.Println("Configuration du tunnel WireGuard (mode Client)...")
				tunnelCfg := tunnel.Config{
					InterfaceName: "wg0",
					PrivateKey:    *wgPrivateKey,
					ListenPort:    51820,
					PeerPublicKey: *wgPeerPublicKey,
					PeerEndpoint:  *wgEndpoint,
					AllowedIPs:    []string{*wgPeerIP},
				}
				if err := tunnel.SetupTunnel(tunnelCfg); err != nil {
					log.Printf("Erreur de configuration du tunnel: %v", err)
				} else {
					setupLocalInterface(*wgLocalIP)
				}
			}

		case proto.TypeConnectReject:
			log.Println("--- CONNEXION REFUSÉE ---")
			dataMap := incoming.Data.(map[string]interface{})
			log.Printf("Raison: %v", dataMap["message"])
		}
	}
}
