package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/blok2ttrpg/charsheet/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv, err := server.New()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	fmt.Printf("Blok2ttrpg Character Sheet running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
