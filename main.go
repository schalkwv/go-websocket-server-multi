package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(port string) http.HandlerFunc {
	var counter int
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade initial GET request to a websocket
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer ws.Close()

		for {
			// Generate a random number
			counter++

			// Construct the response including the port number
			response := map[string]interface{}{
				"port":   port,
				"number": counter,
			}

			// Send the response to the client
			if err := ws.WriteJSON(response); err != nil {
				log.Printf("error: %v", err)
				break
			}

			// Wait for 2 seconds before sending the next number
			time.Sleep(2 * time.Second)
		}
	}
}

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Define ports to run servers on
	ports := []string{"5555", "5556", "5557"}
	for _, port := range ports {
		mux := http.NewServeMux() // Create a new ServeMux for each port
		mux.HandleFunc("/", handleConnections(port))

		go func(port string) {
			log.Printf("Starting server on port %s...\n", port)
			err := http.ListenAndServe(":"+port, mux) // Use the ServeMux for this server
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}(port)
	}

	// Keep the main goroutine running indefinitely
	select {}
}
