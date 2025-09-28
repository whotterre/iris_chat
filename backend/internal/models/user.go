package models

import "github.com/gorilla/websocket"

// User represents a connected user
type User struct {
	ID        string          // Unique identifier for the user
	Conn      *websocket.Conn // WebSocket connection
	SessionID string          // Unique session identifier from client
}
