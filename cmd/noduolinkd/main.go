package main

import (
	"encoding/base64"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/phaladamou/noduolink/api/proto"
	"github.com/phaladamou/noduolink/internal/config"
	"github.com/phaladamou/noduolink/internal/coordination"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	registry = coordination.NewRegistry()
)

func main() {
	configPath := flag.String("config", "config.yaml", "Chemin vers le fichier de configuration")
	flag.Parse()

	log.SetPrefix("[noduolinkd] ")
	log.Println("Démarrage du Coordinateur NoduoLink...")

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Erreur critique de configuration: %v", err)
	}
	log.Printf("Configuration chargée (LogLevel: %s)", cfg.LogLevel)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/ws", handleWebSocket)

	port := ":8080"
	log.Printf("Coordinateur en écoute sur http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Erreur du serveur HTTP: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erreur d'upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Println("Nouvelle connexion WebSocket entrante.")
	var currentClient *coordination.Client

	for {
		var msg proto.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Déconnexion ou erreur de lecture: %v", err)
			break
		}

		switch msg.Type {
		case proto.TypeRegister:
			log.Println("Message de type REGISTER reçu.")
			dataMap, ok := msg.Data.(map[string]interface{})
			if !ok {
				continue
			}
			pubKey, _ := dataMap["public_key"].(string)

			// Validation stricte de la clé publique
			decodedKey, err := base64.StdEncoding.DecodeString(pubKey)
			if err != nil || len(decodedKey) != 32 {
				log.Println("Erreur: clé publique invalide ou mal formée")
				continue
			}

			nodeID := coordination.GenerateNodeID(pubKey)
			currentClient = &coordination.Client{
				NodeID:    nodeID,
				PublicKey: pubKey,
				Conn:      conn,
			}
			registry.Register(currentClient)
			log.Printf("Client enregistré: %s (Total connectés: %d)", nodeID[:8], registry.Count())

			ack := proto.Message{Type: proto.TypeAck, Data: "Bienvenue sur NoduoLink"}
			conn.WriteJSON(ack)

		case proto.TypeConnectRequest:
			dataMap, ok := msg.Data.(map[string]interface{})
			if !ok {
				continue
			}
			targetNodeID, _ := dataMap["target_node_id"].(string)

			targetClient, exists := registry.Get(targetNodeID)
			if !exists {
				errMsg := proto.Message{Type: proto.TypeConnectReject, Data: proto.ConnectResponse{
					TargetNodeID: targetNodeID,
					Accepted:     false,
					Message:      "Nœud cible introuvable ou hors ligne",
				}}
				conn.WriteJSON(errMsg)
				continue
			}

			offer := proto.Message{Type: proto.TypeConnectOffer, Data: proto.ConnectOffer{
				InitiatorNodeID: currentClient.NodeID,
				InitiatorPubKey: currentClient.PublicKey,
			}}
			if err := targetClient.Conn.WriteJSON(offer); err != nil {
				log.Printf("Erreur d'envoi de l'offre: %v", err)
			} else {
				log.Printf("Offre transférée de %s vers %s", currentClient.NodeID[:8], targetNodeID[:8])
			}

		case proto.TypeConnectAccept, proto.TypeConnectReject:
			dataMap, ok := msg.Data.(map[string]interface{})
			if !ok {
				continue
			}
			targetNodeID, _ := dataMap["target_node_id"].(string)

			initiatorClient, exists := registry.Get(targetNodeID)
			if exists {
				initiatorClient.Conn.WriteJSON(msg)
				log.Printf("Réponse (%s) transférée de %s vers %s", msg.Type, currentClient.NodeID[:8], targetNodeID[:8])
			}
		}
	}

	if currentClient != nil {
		registry.Unregister(currentClient.NodeID, conn)
		log.Printf("Client déconnecté: %s (Total: %d)", currentClient.NodeID[:8], registry.Count())
	}
}
