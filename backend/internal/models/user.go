package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

// User represents a connected user
type User struct {
	ID        string          // Unique identifier for the user
	Conn      *websocket.Conn // WebSocket connection
	SessionID string          // Unique session identifier from client
	mu        sync.Mutex      // Write lock for WebSocket connection
}

// Safe WriteJSON to prevent concurrent write panics on WebSocket connection
func (u *User) WriteJSON(v interface{}) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.Conn == nil {
		return nil
	}
	return u.Conn.WriteJSON(v)
}

// Safe WriteMessage to send text message
func (u *User) WriteTextMessage(msg []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.Conn == nil {
		return nil
	}
	return u.Conn.WriteMessage(websocket.TextMessage, msg)
}

// Close safely closes the underlying WebSocket connection
func (u *User) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.Conn != nil {
		return u.Conn.Close()
	}
	return nil
}

