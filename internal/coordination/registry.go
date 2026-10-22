package coordination

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/gorilla/websocket"
)

// Client représente un appareil connecté au coordinateur.
type Client struct {
	NodeID    string
	PublicKey string
	Conn      *websocket.Conn
}

// Registry gère la liste des clients connectés de manière thread-safe.
type Registry struct {
	mu      sync.RWMutex
	clients map[string]*Client
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

// Get retourne un client par son NodeID.
func (r *Registry) Get(nodeID string) (*Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, exists := r.clients[nodeID]
	return client, exists
}

// GetByPrefix cherche un client dont le NodeID commence par le préfixe donné.
// Utile pour l'expérience utilisateur (on n'a pas toujours 64 caractères à copier).
func (r *Registry) GetByPrefix(prefix string) (*Client, bool) {
r.mu.RLock()
defer r.mu.RUnlock()
for id, client := range r.clients {
if len(prefix) <= len(id) && id[:len(prefix)] == prefix {
return client, true
}
}
return nil, false
}
