package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type Template struct {
	Templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func NewTemplateRenderer(e *echo.Echo, paths ...string) {
	tmpl := &template.Template{}
	for i := range paths {
		template.Must(tmpl.ParseGlob(paths[i]))
	}
	t := newTemplate(tmpl)
	e.Renderer = t
}

func newTemplate(templates *template.Template) echo.Renderer {
	return &Template{
		Templates: templates,
	}
}

type messageData struct {
	Port   string
	Number int
}

const (
	// WebSocket endpoint URL
	mainEndPoint = "ws://localhost:5000/"
)

func mainConnection(c chan messageData) {
	// Connect to the WebSocket server
	mainConn, _, err := websocket.DefaultDialer.Dial(mainEndPoint, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer mainConn.Close()

	fmt.Printf("Websocket connected to %s\n", mainEndPoint)
	for {
		// Read message from the server
		_, message, err := mainConn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// marshal message
		var msgData messageData
		err = json.Unmarshal(message, &msgData)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}
		fmt.Println("msgData", msgData)
		c <- msgData
	}
}

func subConnection(c chan messageData, port string) {
	// Connect to the WebSocket server
	addr := fmt.Sprintf("localhost:%s", port)
	url := fmt.Sprintf("ws://%s/", addr)
	mainConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer mainConn.Close()

	fmt.Printf("Websocket connected to %s\n", url)
	for {
		// Read message from the server
		_, message, err := mainConn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// marshal message
		var msgData messageData
		err = json.Unmarshal(message, &msgData)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}
		c <- msgData
	}
}

func main() {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(20),
	)))

	// channel for sending messages to the client
	broadcast := make(chan messageData)

	// connect to main websocket server and start listening for messages
	go mainConnection(broadcast)

	NewTemplateRenderer(e, "public/*.gohtml")

	e.GET("/hello", func(e echo.Context) error {
		res := map[string]interface{}{
			"Name":  "Kiepie",
			"Phone": "8888888",
			"Email": "skyscraper@gmail.com",
		}
		return e.Render(http.StatusOK, "index", res)
	})
	e.GET("/sub/:port", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().WriteHeader(http.StatusOK)

		port := c.Param("port")

		subChannel := make(chan messageData)
		// connect to sub websocket server and
		// start listening for messages
		go subConnection(subChannel, port)
		// listen for messages from the sub websocket server and the main websocket server
		for {
			select {
			// case msg := <-subChannel:
			// 	fmt.Fprintf(c.Response(), "data: %s %d\n\n", msg.Port, msg.Number)
			// 	c.Response().Flush()
			case msg := <-broadcast:
				fmt.Fprintf(c.Response(), "data: %s %d\n\n", msg.Port, msg.Number)
				c.Response().Flush()
			}
		}

	})

	e.Logger.Fatal(e.Start(":4040"))
}
