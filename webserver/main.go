package main

import (
	"fmt"
	"log"
	"net/http"
	"webserver/functionality"
	"webserver/mux"
)

func main() {
	// Initialize the database
	functionality.DB = functionality.InitDB()

	// Set up the router
	router := mux.NewRouter()

	// Define the port
	port := ":8080"
	fmt.Printf("Server is running on http://localhost%s\n", port)

	// Start the server
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
