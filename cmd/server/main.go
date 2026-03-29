package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","message":"GateKeep API - Coming soon!"}`)
	})

	addr := fmt.Sprintf(":%s", port)
	log.Printf("GateKeep API server starting on %s", addr)
	log.Printf("Health check available at http://localhost:%s/api/health", port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
