package main

import "github.com/riverchu/proxy"

const listenPort = "8080"

func main() {
	go proxy.HttpServe(listenPort)

	select {}
}
