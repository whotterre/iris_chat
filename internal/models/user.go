package models

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type UserStatus string

const (
	StatusWaiting UserStatus = "waiting"
	StatusPaired  UserStatus = "paired"
)

type User struct {
	ID     uuid.UUID
	Conn   *websocket.Conn
	Status UserStatus
	RoomID	string
	Mutex 	sync.Mutex
}

func NewUser(conn *websocket.Conn) *User {
	return &User{
		ID: uuid.New(),
		Conn: conn,
		Status: StatusWaiting,
		RoomID: "",
	}
}