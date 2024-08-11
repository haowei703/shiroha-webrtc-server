package main

import (
	"github.com/haowei703/webrtc-server/internal/webrtc"
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		webrtc.StartWebSocketServer()
	}()

	log.Println("Starting servers...")
	wg.Wait()
}
