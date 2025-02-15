package lobby

import (
	"irischat/internal/models"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Lobby struct {
	Users       map[string]*models.User // Map of waiting users
	Mutex       sync.Mutex
	PairingChan chan *models.User // Signals when users are paired
}

func NewLobby() *Lobby {
	return &Lobby{
		Users:       make(map[string]*models.User),
		PairingChan: make(chan *models.User),
	}
}

// Adds user to a lobby
func (l *Lobby) AddUser(user *models.User) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.Users[user.ID] = user
}

// Removes a user from a lobby
func (l *Lobby) RemoveUser(userID string) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	delete(l.Users, userID)
}

// Pair two users from the lobby
func (l *Lobby) PairUsers() {
	for {
		userOne := <-l.PairingChan
		l.Mutex.Lock()

		var userTwo *models.User
		// Find another waiting user in the waiting queue
		for _, u := range l.Users {
			if u.ID != userOne.ID && userOne.Status == models.StatusWaiting {
				userTwo = u
				break
			}
		}

		if userTwo != nil {
			delete(l.Users, userOne.ID)
			delete(l.Users, userTwo.ID)
			l.Mutex.Unlock()

			go pairUsers(userOne, userTwo)
		} else {
			l.Mutex.Unlock()
		}
	}
}

func pairUsers(userOne, userTwo *models.User){
	defer func(){
		userOne.Conn.Close()
		userTwo.Conn.Close()
	}()

	userOne.Conn.WriteMessage(websocket.TextMessage, []byte("You have been paired with " +userTwo.ID))
	userTwo.Conn.WriteMessage(websocket.TextMessage, []byte("You have been paired with " +userOne.ID))

	go handleUserMessages(userOne, userTwo)
    go handleUserMessages(userTwo, userOne)

}


func handleUserMessages(sender, receiver *models.User) {
    for {
        messageType, message, err := sender.Conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }
        if err := receiver.Conn.WriteMessage(messageType, message); err != nil {
            log.Println("Write error:", err)
            break
        }
    }
}
