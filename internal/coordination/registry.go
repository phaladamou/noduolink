package coordination

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/gorilla/websocket"
)

// Client représente un appareil connecté au coordinateur.
type Client struct {
	NodeID    string          // ID calculé à partir de la clé publique
	PublicKey string          // Clé publique en base64
	Conn      *websocket.Conn // La connexion WebSocket
}

// Registry gère la liste des clients connectés de manière thread-safe.
type Registry struct {
	mu      sync.RWMutex
	clients map[string]*Client // La clé de la map est le NodeID
}

// NewRegistry initialise un nouveau registre.
func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[string]*Client),
	}
}

// GenerateNodeID crée un identifiant unique et déterministe basé sur la clé publique.
func GenerateNodeID(publicKeyBase64 string) string {
	hash := sha256.Sum256([]byte(publicKeyBase64))
	return hex.EncodeToString(hash[:])
}

// Register ajoute ou remplace un client dans le registre.
func (r *Registry) Register(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Si un ancien client avec le même ID existe, on le ferme proprement
	if oldClient, exists := r.clients[client.NodeID]; exists {
		// Sécurité : on vérifie que la connexion n'est pas nil (utile pour les tests)
		if oldClient.Conn != nil {
			oldClient.Conn.Close()
		}
	}

	r.clients[client.NodeID] = client
}

// Unregister supprime un client du registre.
func (r *Registry) Unregister(nodeID string, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if client, exists := r.clients[nodeID]; exists && client.Conn == conn {
		delete(r.clients, nodeID)
	}
}

// Count retourne le nombre de clients connectés.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}
