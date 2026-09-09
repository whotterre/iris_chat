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

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	frontendDir := findFrontendDir()
	log.Printf("Serving frontend static files from: %s", frontendDir)

	fs := http.FileServer(http.Dir(frontendDir))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handlers.HandleWebSocket)
	http.HandleFunc("/health", corsMiddleware(handlers.HandleHealth))
	http.HandleFunc("/stats", corsMiddleware(handlers.HandleStats))

	addr := ":" + port
	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
