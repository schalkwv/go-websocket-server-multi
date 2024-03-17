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

func main() {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(20),
	)))

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
		addr := fmt.Sprintf("localhost:%s", port)
		url := fmt.Sprintf("ws://%s/", addr)

		// Connect to the WebSocket server
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
			select {
			case <-c.Request().Context().Done():
				return nil
			default:
			}
			fmt.Fprintf(c.Response(), "data: main %d\n\n", msgData.Number)
			c.Response().Flush()

			// Log the message to the terminal
			// fmt.Printf("Message: %s\n", message)
		}
		return nil
	})

	e.Logger.Fatal(e.Start(":4040"))
}
