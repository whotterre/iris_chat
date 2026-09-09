package models

import (
	"fmt"
	"log"
	"sync"
)

type Lobby struct {
	WaitingUsers []*User          // Users waiting for a match
	Rooms        map[string]*Room // Active chat rooms
	Mutex        sync.Mutex       // Synchronization for concurrent access
}

// GetTotalConnections returns total active connections (waiting + in rooms)
func (l *Lobby) GetTotalConnections() int {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	return l.GetTotalConnectionsUnlocked()
}

func (l *Lobby) GetTotalConnectionsUnlocked() int {
	total := len(l.WaitingUsers)
	for _, room := range l.Rooms {
		total += len(room.Users)
	}
	return total
}

// BroadcastConnectionCount updates all connected clients with current online count
func (l *Lobby) BroadcastConnectionCount() {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	total := l.GetTotalConnectionsUnlocked()
	log.Printf("[CONNECTIONS] Total active connections: %d\n", total)

	msg := map[string]string{
		"sender":  "Server",
		"type":    "connections",
		"message": fmt.Sprintf("Total active connections: %d", total),
		"count":   fmt.Sprintf("%d", total),
	}

	// Broadcast to waiting users
	for _, u := range l.WaitingUsers {
		go u.WriteJSON(msg)
	}

	// Broadcast to users in rooms
	for _, room := range l.Rooms {
		for _, u := range room.Users {
			go u.WriteJSON(msg)
		}
	}
}

// RemoveWaitingUser removes a user from waiting list if present
func (l *Lobby) RemoveWaitingUserUnlocked(userID string) bool {
	for i, u := range l.WaitingUsers {
		if u.ID == userID {
			l.WaitingUsers = append(l.WaitingUsers[:i], l.WaitingUsers[i+1:]...)
			return true
		}
	}
	return false
}