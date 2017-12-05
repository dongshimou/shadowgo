package main

import (
	"./test"
	"time"
)

func main() {
	go test.SocksProxy()
	go test.ProxyServerTest()

	for {
		time.Sleep(time.Second)
	}
}
