package main

import (
	"sync"
	"time"

	"github.com/ntekim/grpc-cli-quiz/server"
)

func main() {
	// Use a WaitGroup to synchronize the server and CLI
	var wg sync.WaitGroup

	// Start gRPC server in a goroutine
	wg.Add(1)
	go server.NewServer(&wg)

	// Allow server to start before the client attempts to connect
	time.Sleep(2 * time.Second)

	// Run the CLI in the main goroutine
	StartCLI()

	wg.Wait() // Wait for the server to finish (optional: server won't exit unless killed)
}
