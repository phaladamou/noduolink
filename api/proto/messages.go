package proto

// MessageType définit le type de message échangé.
type MessageType string

const (
	TypeRegister MessageType = "register"
	TypeAck      MessageType = "ack"

	// M2 : Signalisation
	TypeConnectRequest MessageType = "connect_request"
	TypeConnectOffer   MessageType = "connect_offer"
	TypeConnectAccept  MessageType = "connect_accept"
	TypeConnectReject  MessageType = "connect_reject"
)

// Message est l'enveloppe standard de communication.
type Message struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// RegisterRequest est envoyé par le client au démarrage.
type RegisterRequest struct {
	NodeID    string `json:"node_id"`
	PublicKey string `json:"public_key"`
}

// ConnectRequest est envoyé par le client initiateur.
type ConnectRequest struct {
	TargetNodeID string `json:"target_node_id"`
}

// ConnectOffer est envoyé par le serveur au nœud cible.
type ConnectOffer struct {
	InitiatorNodeID string `json:"initiator_node_id"`
	InitiatorPubKey string `json:"initiator_pub_key"`
}

// ConnectResponse est utilisé pour accepter ou refuser.
type ConnectResponse struct {
	TargetNodeID string `json:"target_node_id"`
	Accepted     bool   `json:"accepted"`
	Message      string `json:"message,omitempty"`
}
