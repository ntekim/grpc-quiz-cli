package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		err := StartCLI(ctx)
		if err != nil {
			log.Println("Error executing CLI:", err)
			cancel()
		}
	}()

	<-ctx.Done()

	log.Println("All goroutines finished. Exiting.")
	os.Exit(0)
}
