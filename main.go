package main

import (
	"./test"
	"time"
)
func main(){
	go test.TcpProxy()
	go test.ProxyServerTest()

	for{
		time.Sleep(time.Second*10);
	}
	}
