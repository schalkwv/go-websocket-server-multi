package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("failed to start the app: %v", err)
	}
}

func run() error {
	ctx := context.Background()
	go func() {
		// listening for the shutdown signal
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		<-quit

		log.Println("shutting down the server...")

		err := e.Shutdown(context.Background()) // TODO: consider shutdown timeout
		if err != nil {
			log.Println("error shutting down the server: ", err)
		}

		close(done)
	}()

