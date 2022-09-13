package main

import (
	"flag"

	"github.com/riverchu/proxy"
)

var (
	listenPort int
)

func init() {
	listenPort = *flag.Int("port", 8080, "listen port")
}

func main() {
	go proxy.HttpServe(listenPort)

	select {}
}
