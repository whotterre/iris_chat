package models
// Message represents a chat message
type Message struct {
	SenderID  string // ID of the sender
	Content   string // Message content
	Timestamp string // Timestamp of the message
}