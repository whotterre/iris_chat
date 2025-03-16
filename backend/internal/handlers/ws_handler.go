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

// HandleWebSocket manages user connections, room creation, and messaging
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	// Generate a unique user ID
	userID := uuid.New().String()
	user := &models.User{
		ID:   userID,
		Conn: conn,
	}

	lobby.Mutex.Lock()
	lobby.WaitingUsers = append(lobby.WaitingUsers, user)
	room := assignRoom()
	lobby.Mutex.Unlock()

	// Notify the user of their status
	if room == nil || room.ID == "waiting" {
		initialMsg := map[string]string{
			"sender":  "Server",
			"message": "Waiting for another user...",
		}
		jsonMsg, _ := json.Marshal(initialMsg)
		conn.WriteMessage(websocket.TextMessage, jsonMsg)
	} else {
		// Notify both users in the room
		for _, u := range room.Users {
			msg := map[string]string{
				"sender":  "Server",
				"message": fmt.Sprintf("Connected to room %s", room.ID),
			}
			jsonMsg, _ := json.Marshal(msg)
			u.Conn.WriteMessage(websocket.TextMessage, jsonMsg)

			// Start listening for messages from each user in their own goroutine
			go listenForMessages(u, room)
		}
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
		lobby.WaitingUsers = lobby.WaitingUsers[2:]

		fmt.Println("[ROOM CREATED]", roomID)
		return room
	}

	// No available pair yet, user must wait
	return &models.Room{ID: "waiting"}
}

// Listens for messages from a user and broadcasts them to their room
func listenForMessages(user *models.User, room *models.Room) {
	for {
		_, msg, err := user.Conn.ReadMessage()
		if err != nil {
			fmt.Println("[DISCONNECT] User disconnected:", user.ID, "Error:", err)
			removeUserFromRoom(user.ID)
			break
		}
		
		var messageData map[string]string
		if err := json.Unmarshal(msg, &messageData); err != nil {
			fmt.Println("[ERROR] Failed to parse message:", err)
			continue
		}

		if messageData["type"] == "leave" {
			removeUserFromRoom(user.ID)
			user.Conn.Close()
			break
		}

		fmt.Println("[MESSAGE RECEIVED] From:", user.ID, "Message:", string(msg))
		broadcastMessage(room, user.ID, msg)
	}
}

// Broadcasts messages within a room
func broadcastMessage(room *models.Room, senderID string, msg []byte) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	var parsedMsg map[string]string
	if err := json.Unmarshal(msg, &parsedMsg); err != nil {
		fmt.Println("[ERROR] Failed to parse incoming message:", err)
		return
	}

	messageData := map[string]string{
		"sender":  senderID,
		"message": parsedMsg["message"],
	}

	jsonMsg, _ := json.Marshal(messageData)

	for _, user := range room.Users {
		if user.ID != senderID { 
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
			fmt.Println("[USER LEFT] User", userID, "left room", roomID)

			// Notify the other user in the room
			for _, remainingUser := range room.Users {
				disconnectMsg := map[string]string{
					"sender":  "Server",
					"message": "The other user has disconnected. You are now alone in the room.",
				}
				jsonMsg, _ := json.Marshal(disconnectMsg)
				remainingUser.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
			}

			// Remove empty rooms
			if len(room.Users) == 0 {
				delete(lobby.Rooms, roomID)
				fmt.Println("[ROOM DELETED] Empty room", roomID, "removed")
			}
			break
		}
	}
}
