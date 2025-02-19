package main

import (
	"fmt"
	"irischat/backend/internal/handlers"
	"net/http"
)

func main() {
	// Routes
	http.HandleFunc("/ws", handlers.HandleWebSocket)

	fmt.Println("Server started on :4000")
	http.ListenAndServe(":4000", nil)
}
