package main

import (
	"sync"
	"time"

	"github.com/ntekim/grpc-cli-quiz/server"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go server.NewServer(&wg)

	time.Sleep(2 * time.Second)

	StartCLI()

	wg.Wait()
}
