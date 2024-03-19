package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	appVersion  string
	mainChannel chan messageData
}

// Validator is a custom validator for Echo.
type Validator struct {
	validator *validator.Validate
}

// Validate validates the request according to the required tags.
// Returns HTTPError if the required parameter is missing in the request.
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func NewServer(appVersion string) *Server {
	broadcast := make(chan messageData)
	go mainConnection(broadcast)
	return &Server{
		appVersion:  appVersion,
		mainChannel: broadcast,
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

func (s *Server) Router() *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Validator = &Validator{validator: validator.New()}
	e.Static("/static", "static")

	e.GET("/version", s.versionHandler)
	e.GET("/", s.rootHandler)
	e.GET("/content", s.contentHandler)
	// channel for sending messages to the client
	// broadcast := make(chan messageData)
	//
	// // connect to main websocket server and start listening for messages
	// go mainConnection(broadcast)

	// e.GET("/sub/:port", s.subHandler)

	return e
}

func (s *Server) versionHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.appVersion)
}

type account struct {
	Name string `json:"name" validate:"required"`
	Port int    `json:"port" validate:"required"`
}

type accountList map[string]account

func (s *Server) subHandler(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().WriteHeader(http.StatusOK)

	// port := c.Param("port")
	//
	// subChannel := make(chan messageData)
	// // connect to sub websocket server and
	// // start listening for messages
	// go subConnection(subChannel, port)
	// listen for messages from the sub websocket server and the main websocket server
	for {
		select {
		// case msg := <-subChannel:
		// 	fmt.Fprintf(c.Response(), "data: %s %d\n\n", msg.Port, msg.Number)
		// 	c.Response().Flush()
		case msg := <-s.mainChannel:
			fmt.Println("msg", msg)
			fmt.Fprintf(c.Response(), "data: %s %d\n\n", msg.Port, msg.Number)
			c.Response().Flush()
		}
	}
}

func (s *Server) contentHandler(c echo.Context) error {
	tmpl := getTemplate()
	accounts := accountList{
		"account1": {
			Name: "Account 1",
			Port: 5001,
		},
		"account2": {
			Name: "Account 2",
			Port: 5002,
		},
		"account3": {
			Name: "Account 3",
			Port: 5003,
		},
	}
	err := tmpl.ExecuteTemplate(c.Response().Writer, "Content",
		map[string]any{
			"Accounts": accounts,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (s *Server) rootHandler(c echo.Context) error {
	tmpl := getTemplate()

	accounts := accountList{
		"account1": {
			Name: "Account 1",
			Port: 5001,
		},
		"account2": {
			Name: "Account 2",
			Port: 5002,
		},
		"account3": {
			Name: "Account 3",
			Port: 5003,
		},
	}

	err := tmpl.ExecuteTemplate(c.Response().Writer, "Base",
		map[string]any{
			"Accounts": accounts,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}
