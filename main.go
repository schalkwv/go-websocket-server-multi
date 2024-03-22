package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type messageData struct {
	Port   string
	Number int
}

// connection wraps the websocket connection and the send channel.
type connection struct {
	ws   *websocket.Conn
	send chan messageData
}

// runWriter listens on the send channel and sends messages to the client.
func (c *connection) runWriter() {
	// for message := range c.send {
	// 	if err := c.ws.WriteJSON(map[string]interface{}{"number": message}); err != nil {
	// 		break
	// 	}
	// }
	for msg := range c.send {
		// Convert the messageData struct to JSON for sending
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}
		if err := c.ws.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
			log.Printf("Failed to send message: %v", err)
			break
		}
	}
	c.ws.Close()
}

func serveWs(port string, pool *sync.Pool, addConn chan *connection, removeConn chan *connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}

		conn := pool.Get().(*connection)
		conn.ws = ws
		conn.send = make(chan messageData, 256)

		addConn <- conn

		// Start writer goroutine
		go conn.runWriter()

		// Wait for connection to close
		_, _, err = ws.ReadMessage()
		if err != nil {
			removeConn <- conn
			pool.Put(conn)
			ws.Close()
		}
	}
}

type serverType struct {
	port  string
	delay time.Duration
}

func main() {
	servers := []serverType{
		{"5000", 1000},
		{"5001", 2000},
		{"5002", 4000},
		{"5003", 8000},
		{"5004", 16000},
	}
	// ports := []string{"5555", "5556", "5557"}
	inc := 0
	for _, server := range servers {
		inc++
		addConn := make(chan *connection)
		removeConn := make(chan *connection)
		conns := make(map[*connection]bool)
		counter := 0

		// Connection pool to reuse connection objects
		pool := &sync.Pool{
			New: func() interface{} {
				return &connection{}
			},
		}

		go func() {
			for {
				select {
				case conn := <-addConn:
					conns[conn] = true
				case conn := <-removeConn:
					if _, ok := conns[conn]; ok {
						delete(conns, conn)
						close(conn.send)
					}
				}
			}
		}()

		go func(port string, inc int, delay time.Duration) {
			for {
				time.Sleep(delay * time.Millisecond)
				counter += inc
				msg := messageData{
					Port:   port,
					Number: counter,
				}
				for conn := range conns {
					select {
					case conn.send <- msg:
					default:
						close(conn.send)
						delete(conns, conn)
					}
				}
			}
		}(server.port, inc, server.delay)

		mux := http.NewServeMux()
		mux.HandleFunc("/", serveWs(server.port, pool, addConn, removeConn))

		go func(port string) {
			log.Printf("Starting server on port %s...\n", port)
			if err := http.ListenAndServe(":"+port, mux); err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}(server.port)
	}

	select {}
}
