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

	"github.com/kelseyhightower/envconfig"

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
