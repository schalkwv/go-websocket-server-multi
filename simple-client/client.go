package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	// Define and parse the port command-line flag
	port := flag.String("port", "5555", "the port to connect to")
	flag.Parse()

	// Construct the WebSocket URL
	addr := fmt.Sprintf("localhost:%s", *port)
	url := fmt.Sprintf("ws://%s/", addr)

	// Connect to the WebSocket server
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	fmt.Printf("Connected to %s\n", url)

	for {
		// Read message from the server
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		// Log the message to the terminal
		fmt.Printf("Message: %s\n", message)
	}
}
