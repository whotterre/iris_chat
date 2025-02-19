package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"irischat/backend/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

// HandleWebSocket manages everything: user connections, room creation, and messaging
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Generate a unique user ID
	userID := uuid.New().String()
	user := &models.User{
		ID:   userID,
		Conn: conn,
	}

	lobby.Mutex.Lock()
	lobby.WaitingUsers = append(lobby.WaitingUsers, user)
	room := assignRoom() // Now this function will wait for a second user
	lobby.Mutex.Unlock()

	initialMsg := map[string]string{
		"sender": "Server",
	}

	if room == nil {
		initialMsg["message"] = "Waiting for another user..."
	} else {
		initialMsg["message"] = fmt.Sprintf("Connected as %s in room %s", userID, room.ID)
	}

	jsonMsg, _ := json.Marshal(initialMsg)
	conn.WriteMessage(websocket.TextMessage, jsonMsg)

	// Notify users in the room
	for _, u := range room.Users {
		u.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("You are in room %s", room.ID)))
	}

	// Handle incoming messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("[DISCONNECT] User disconnected:", userID, "Error:", err)
			removeUserFromRoom(userID)
			break
		}
		fmt.Println("[MESSAGE RECEIVED] From:", userID, "Message:", string(msg))

		broadcastMessage(room, userID, msg)
	}
}

// Assigns users to a room or creates a new one
func assignRoom() *models.Room {
	if len(lobby.WaitingUsers) >= 2 {
		player1 := lobby.WaitingUsers[0]
		player2 := lobby.WaitingUsers[1]
		roomID := "Room-" + uuid.New().String()

		room := &models.Room{
			ID:    roomID,
			Users: map[string]*models.User{player1.ID: player1, player2.ID: player2},
		}

		lobby.Rooms[roomID] = room
		fmt.Print(len(lobby.WaitingUsers))
		lobby.WaitingUsers = lobby.WaitingUsers[2:]

		fmt.Println("Room created:", roomID)
		return room
	}

	// If no room yet, return a placeholder (won't be used until a second user joins)
	return &models.Room{ID: "waiting"}
}

// Broadcasts messages within a room
func broadcastMessage(room *models.Room, senderID string, msg []byte) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	messageData := map[string]string{
		"sender":  senderID,
		"message": string(msg),
	}
	
	jsonMsg, _ := json.Marshal(messageData)

	for _, user := range room.Users {
		if user.ID != senderID { // Don't send the message back to the sender
			err := user.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
			if err != nil {
				fmt.Println("[ERROR] Failed to send message to user:", user.ID, err)
			}
		}
	}
}

// Removes a user when they disconnect
func removeUserFromRoom(userID string) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	for roomID, room := range lobby.Rooms {
		if _, exists := room.Users[userID]; exists {
			delete(room.Users, userID)
			fmt.Println("User", userID, "left room", roomID)

			// Remove empty rooms
			if len(room.Users) == 0 {
				delete(lobby.Rooms, roomID)
			}
			break
		}
	}
}
