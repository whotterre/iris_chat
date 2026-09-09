package main

import (
	"log"
	"net/http"
	"os"

	"irischat/backend/internal/handlers"
)

func findFrontendDir() string {
	paths := []string{"./frontend", "../frontend", "../../frontend"}
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return "./frontend"
}

func main() {
	frontendDir := findFrontendDir()
	log.Printf("Serving frontend from: %s", frontendDir)

	fs := http.FileServer(http.Dir(frontendDir))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handlers.HandleWebSocket)
	http.HandleFunc("/health", handlers.HandleHealth)
	http.HandleFunc("/stats", handlers.HandleStats)

	addr := "0.0.0.0:4000"
	log.Printf("Server running on http://localhost:4000")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
