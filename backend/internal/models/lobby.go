package models

import "sync"

type Lobby struct {
	WaitingUsers []*User   // Users waiting for a match
	Rooms        map[string]*Room // Active chat rooms
	Mutex        sync.Mutex // Synchronization for concurrent access
}