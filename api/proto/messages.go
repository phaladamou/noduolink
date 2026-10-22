package proto

// MessageType définit le type de message échangé.
type MessageType string

const (
	// TypeRegister est envoyé par le client pour s'enregistrer.
	TypeRegister MessageType = "register"
	// TypeAck est envoyé par le serveur pour confirmer.
	TypeAck MessageType = "ack"
)

// Message est l'enveloppe standard de communication entre le client et le coordinateur.
type Message struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// RegisterRequest est le message envoyé par le client au démarrage.
type RegisterRequest struct {
	NodeID    string `json:"node_id"`    // Un ID généré localement ou basé sur la clé publique
	PublicKey string `json:"public_key"` // La clé publique en base64
}
