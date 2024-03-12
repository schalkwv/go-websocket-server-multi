package main

import (
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

// connection wraps the websocket connection and the send channel.
type connection struct {
	ws   *websocket.Conn
	send chan int
}

// runWriter listens on the send channel and sends messages to the client.
func (c *connection) runWriter() {
	for message := range c.send {
		if err := c.ws.WriteJSON(map[string]interface{}{"number": message}); err != nil {
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
		conn.send = make(chan int, 256)

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

func main() {
	ports := []string{"5555", "5556", "5557"}
	for _, port := range ports {
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

		go func(port string) {
			for {
				time.Sleep(2 * time.Second)
				counter++
				for conn := range conns {
					select {
					case conn.send <- counter:
					default:
						close(conn.send)
						delete(conns, conn)
					}
				}
			}
		}(port)

		mux := http.NewServeMux()
		mux.HandleFunc("/", serveWs(port, pool, addConn, removeConn))

		go func(port string) {
			log.Printf("Starting server on port %s...\n", port)
			if err := http.ListenAndServe(":"+port, mux); err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}(port)
	}

	select {}
}
