package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/phaladamou/noduolink/api/proto"
	"github.com/phaladamou/noduolink/internal/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Autoriser toutes les origines pour le développement initial
	},
}

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

	// Route HTTP de base pour vérifier que le serveur tourne
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Route WebSocket pour les agents
	http.HandleFunc("/ws", handleWebSocket)

	port := ":8080"
	log.Printf("Coordinateur en écoute sur http://localhost%s (WebSocket: /ws)", port)

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

	// Boucle de lecture des messages
	for {
		var msg proto.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Déconnexion ou erreur de lecture: %v", err)
			break
		}

		switch msg.Type {
		case proto.TypeRegister:
			log.Println("Message de type REGISTER reçu.")
			// On renvoie un ACK au client
			ack := proto.Message{Type: proto.TypeAck, Data: "Bienvenue sur NoduoLink"}
			if err := conn.WriteJSON(ack); err != nil {
				log.Printf("Erreur d'envoi de l'ACK: %v", err)
			}
		default:
			log.Printf("Type de message inconnu: %s", msg.Type)
		}
	}
}
