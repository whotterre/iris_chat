package main

import (
	"fmt"
	"irischat/backend/internal/handlers"
	"net/http"
)

func main() {
	// Routes
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	addr := "0.0.0.0:4000"
	fmt.Println("Server started on :4000")
	http.ListenAndServe(addr, nil)
}
