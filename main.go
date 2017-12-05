package main

import (
	"./test"
	"time"
)

func main() {
	go test.ClientListen()
	go test.ServerListen()

	for {
		time.Sleep(time.Second * 10)
	}
}
