package models

import "sync"


type Room struct {
	ID      string         // Unique identifier for the room
	Users   map[string]*User // Connected users
	Mutex   sync.Mutex     // Synchronization for concurrent access
	LogFile string         // Log file for messages
}