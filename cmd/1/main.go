package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"

	"github.com/schalkwv/go-websocket-server-multi/config"
	"github.com/schalkwv/go-websocket-server-multi/internal/handler"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("failed to start the app: %v", err)
	}
}

func run() error {
	// ctx := context.Background()
	var cfg config.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return fmt.Errorf("invalid env config: %w", err)
	}

	server := handler.NewServer("1.0.0")
	router := server.Router()

	router.GET("/events", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().WriteHeader(http.StatusOK)

		// Create a channel to send data
		dataCh := make(chan string)

		// Create a context for handling client disconnection
		// _, cancel := context.WithCancel(r.Context())
		// defer cancel()

		// Send data to the client
		go func() {
			for data := range dataCh {
				fmt.Fprintf(c.Response(), "data: %s\n\n", data)
				c.Response().Flush()
			}
		}()

		// Simulate sending data periodically
		for {
			dataCh <- time.Now().Format(time.TimeOnly)
			time.Sleep(800 * time.Millisecond)
		}

	})
	// http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
	// 	// Set headers for SSE
	// 	w.Header().Set("Content-Type", "text/event-stream")
	// 	w.Header().Set("Cache-Control", "no-cache")
	// 	w.Header().Set("Connection", "keep-alive")
	//
	// 	// Create a channel to send data
	// 	dataCh := make(chan string)
	//
	// 	// Create a context for handling client disconnection
	// 	_, cancel := context.WithCancel(r.Context())
	// 	defer cancel()
	//
	// 	// Send data to the client
	// 	go func() {
	// 		for data := range dataCh {
	// 			fmt.Fprintf(w, "data: %s\n\n", data)
	// 			w.(http.Flusher).Flush()
	// 		}
	// 	}()
	//
	// 	// Simulate sending data periodically
	// 	for {
	// 		dataCh <- time.Now().Format(time.TimeOnly)
	// 		time.Sleep(1 * time.Second)
	// 	}
	// })

	done := make(chan struct{}) // closed when the server shutdown is complete
	go func() {
		// listening for the shutdown signal
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		<-quit

		log.Println("shutting down the server...")

		err := router.Shutdown(context.Background()) // TODO: consider shutdown timeout
		if err != nil {
			log.Println("error shutting down the server: ", err)
		}

		close(done)
	}()

	err = router.Start(":" + cfg.Port)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	log.Println("server gracefully stopped")

	return nil
}
