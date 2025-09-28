package models

import "github.com/gorilla/websocket"

// User represents a connected user
 type User struct {
	ID   string          // Unique identifier for the user
	Conn *websocket.Conn // WebSocket connection
	IP string
}
