package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"irischat/backend/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	lobby = &models.Lobby{
		WaitingUsers: []*models.User{},
		Rooms:        make(map[string]*models.Room),
		Mutex:        sync.Mutex{},
	}
)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	userID := uuid.New().String()
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = userID
	}

	user := &models.User{
		ID:        userID,
		Conn:      conn,
		SessionID: sessionID,
	}

	log.Printf("User connected: %s (Session: %s)", userID, sessionID)

	lobby.Mutex.Lock()
	evictSessionUnlocked(sessionID)
	lobby.WaitingUsers = append(lobby.WaitingUsers, user)
	tryMatchmakingUnlocked()
	lobby.Mutex.Unlock()

	lobby.BroadcastConnectionCount()
	sendUserStatus(user)

	go pingLoop(user)
	go readLoop(user)
}

// evictSessionUnlocked purges any previous stale connection for the same session ID
func evictSessionUnlocked(sessionID string) {
	if sessionID == "" {
		return
	}

	// Evict from WaitingUsers
	for i := len(lobby.WaitingUsers) - 1; i >= 0; i-- {
		u := lobby.WaitingUsers[i]
		if u.SessionID == sessionID {
			log.Printf("Evicting stale waiting connection for session: %s (User ID: %s)", sessionID, u.ID)
			u.Close()
			lobby.WaitingUsers = append(lobby.WaitingUsers[:i], lobby.WaitingUsers[i+1:]...)
		}
	}

	// Evict from Rooms
	for roomID, room := range lobby.Rooms {
		for uID, u := range room.Users {
			if u.SessionID == sessionID {
				log.Printf("Evicting stale room connection for session: %s (User ID: %s in Room %s)", sessionID, uID, roomID)
				u.Close()
				delete(room.Users, uID)

				for partnerID, partner := range room.Users {
					delete(room.Users, partnerID)
					lobby.WaitingUsers = append(lobby.WaitingUsers, partner)
					go partner.WriteJSON(map[string]string{
						"sender":  "Server",
						"type":    "partner_left",
						"message": "Stranger reconnected. Searching for a new match...",
					})
				}

				if len(room.Users) == 0 {
					delete(lobby.Rooms, roomID)
				}
				break
			}
		}
	}
}

func sendUserStatus(user *models.User) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	room := findUserRoomUnlocked(user.ID)
	if room != nil {
		_ = user.WriteJSON(map[string]string{
			"sender":  "Server",
			"type":    "status",
			"status":  "connected",
			"message": fmt.Sprintf("Connected to room %s", room.ID),
			"room_id": room.ID,
		})
	} else {
		_ = user.WriteJSON(map[string]string{
			"sender":  "Server",
			"type":    "status",
			"status":  "waiting",
			"message": "Waiting for partner...",
		})
	}
}

func pingLoop(user *models.User) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := user.WriteTextMessage([]byte{}); err != nil {
			return
		}
	}
}

func readLoop(user *models.User) {
	defer cleanupUser(user.ID)

	user.Conn.SetReadLimit(maxMessageSize)
	_ = user.Conn.SetReadDeadline(time.Now().Add(pongWait))
	user.Conn.SetPongHandler(func(string) error {
		_ = user.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := user.Conn.ReadMessage()
		if err != nil {
			log.Printf("User %s disconnect: %v", user.ID, err)
			break
		}

		_ = user.Conn.SetReadDeadline(time.Now().Add(pongWait))

		var messageData map[string]string
		if err := json.Unmarshal(msg, &messageData); err != nil {
			log.Printf("Failed to parse message from %s: %v", user.ID, err)
			continue
		}

		switch messageData["type"] {
		case "leave":
			return

		case "find_next":
			handleFindNext(user)

		case "typing":
			handleTyping(user, messageData["is_typing"])

		default:
			handleChatMessage(user, messageData["message"])
		}
	}
}

func handleTyping(user *models.User, isTyping string) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	room := findUserRoomUnlocked(user.ID)
	if room == nil {
		return
	}

	payload := map[string]string{
		"sender":    user.ID,
		"type":      "typing",
		"is_typing": isTyping,
	}

	for partnerID, partner := range room.Users {
		if partnerID != user.ID {
			go partner.WriteJSON(payload)
		}
	}
}

func handleChatMessage(user *models.User, content string) {
	if content == "" {
		return
	}

	lobby.Mutex.Lock()
	room := findUserRoomUnlocked(user.ID)
	if room == nil {
		lobby.Mutex.Unlock()
		_ = user.WriteJSON(map[string]string{
			"sender":  "Server",
			"type":    "error",
			"message": "Not connected to a room.",
		})
		return
	}

	payload := map[string]string{
		"sender":    user.ID,
		"type":      "chat",
		"message":   content,
		"timestamp": time.Now().Format("15:04"),
	}

	for partnerID, partner := range room.Users {
		if partnerID != user.ID {
			go partner.WriteJSON(payload)
		}
	}
	lobby.Mutex.Unlock()
}

func handleFindNext(user *models.User) {
	lobby.Mutex.Lock()

	room := findUserRoomUnlocked(user.ID)
	if room != nil {
		delete(room.Users, user.ID)

		for partnerID, partner := range room.Users {
			go partner.WriteJSON(map[string]string{
				"sender":  "Server",
				"type":    "partner_left",
				"message": "Stranger has left. Searching for a new match...",
			})
			delete(room.Users, partnerID)
			lobby.WaitingUsers = append(lobby.WaitingUsers, partner)
			go partner.WriteJSON(map[string]string{
				"sender":  "Server",
				"type":    "status",
				"status":  "waiting",
				"message": "Waiting for partner...",
			})
		}

		if len(room.Users) == 0 {
			delete(lobby.Rooms, room.ID)
		}
	}

	alreadyWaiting := false
	for _, u := range lobby.WaitingUsers {
		if u.ID == user.ID {
			alreadyWaiting = true
			break
		}
	}
	if !alreadyWaiting {
		lobby.WaitingUsers = append(lobby.WaitingUsers, user)
	}

	tryMatchmakingUnlocked()
	lobby.Mutex.Unlock()

	lobby.BroadcastConnectionCount()
	sendUserStatus(user)
}

func cleanupUser(userID string) {
	lobby.Mutex.Lock()

	removedFromWaiting := lobby.RemoveWaitingUserUnlocked(userID)

	var affectedPartner *models.User
	for roomID, room := range lobby.Rooms {
		if user, exists := room.Users[userID]; exists {
			user.Close()
			delete(room.Users, userID)

			for partnerID, partner := range room.Users {
				affectedPartner = partner
				delete(room.Users, partnerID)
				lobby.WaitingUsers = append(lobby.WaitingUsers, partner)
			}

			if len(room.Users) == 0 {
				delete(lobby.Rooms, roomID)
			}
			break
		}
	}

	if removedFromWaiting {
		log.Printf("Removed user %s from waiting list", userID)
	}

	if affectedPartner != nil {
		tryMatchmakingUnlocked()
	}

	lobby.Mutex.Unlock()

	if affectedPartner != nil {
		go affectedPartner.WriteJSON(map[string]string{
			"sender":  "Server",
			"type":    "partner_left",
			"message": "Stranger disconnected. Searching for a new match...",
		})
		sendUserStatus(affectedPartner)
	}

	lobby.BroadcastConnectionCount()
}

func tryMatchmakingUnlocked() {
	if len(lobby.WaitingUsers) < 2 {
		return
	}

	i := 0
	for i < len(lobby.WaitingUsers) {
		matched := false
		for j := i + 1; j < len(lobby.WaitingUsers); j++ {
			u1 := lobby.WaitingUsers[i]
			u2 := lobby.WaitingUsers[j]

			if u1.SessionID != u2.SessionID || u1.ID != u2.ID {
				roomID := "Room-" + uuid.New().String()
				room := &models.Room{
					ID:    roomID,
					Users: map[string]*models.User{u1.ID: u1, u2.ID: u2},
				}
				lobby.Rooms[roomID] = room

				lobby.WaitingUsers = append(lobby.WaitingUsers[:j], lobby.WaitingUsers[j+1:]...)
				lobby.WaitingUsers = append(lobby.WaitingUsers[:i], lobby.WaitingUsers[i+1:]...)

				log.Printf("Matched room %s: %s & %s", roomID, u1.ID, u2.ID)

				msg := map[string]string{
					"sender":  "Server",
					"type":    "status",
					"status":  "connected",
					"message": fmt.Sprintf("Connected to room %s", roomID),
					"room_id": roomID,
				}
				go u1.WriteJSON(msg)
				go u2.WriteJSON(msg)

				matched = true
				break
			}
		}
		if !matched {
			i++
		}
	}
}

func findUserRoomUnlocked(userID string) *models.Room {
	for _, room := range lobby.Rooms {
		if _, exists := room.Users[userID]; exists {
			return room
		}
	}
	return nil
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "iris-chat"})
}

func HandleStats(w http.ResponseWriter, r *http.Request) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"active_connections": lobby.GetTotalConnectionsUnlocked(),
		"waiting_users":      len(lobby.WaitingUsers),
		"active_rooms":        len(lobby.Rooms),
	})
}
