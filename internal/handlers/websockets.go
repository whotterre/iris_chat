package handlers

import (
	"irischat/internal/models"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)
// Set up upgrader to upgrade HTTP requests to WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}, //Allow all connections
}

func HandleWebSockets(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Create a new user instance
	currentUser := models.NewUser(conn)
	log.Printf("New user connected %s", currentUser.ID)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Recieved: %s", message)

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("Write error:", err)
			break
		}
	}

}
