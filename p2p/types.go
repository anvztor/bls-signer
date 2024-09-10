package p2p

type Message struct {
	MessageType MessageType `json:"id"`
	Content     string      `json:"content"`
}

type HeartbeatMessage struct {
	PeerID    string `json:"peer_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"ts"`
}

type SignatureMessage struct {
	PeerID    string
	SignDoc   []byte
	Signature []byte
	PublicKey []byte
}

type MessageType int

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeSignature
	MessageTypeResponse
)

var dbDir = "/app/db"
