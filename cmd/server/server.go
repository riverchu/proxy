package main

import (
	"flag"
	"os"

	"github.com/riverchu/pkg/log"
	"github.com/riverchu/proxy"
)

var (
	listenPort int
)

func init() {
	listenPort = *flag.Int("port", 8080, "listen port")
}

func main() {
	log.Info("this is a proxy server")
	if p := os.Getenv("http_proxy"); p != "" {
		log.Info("detect http proxy: %s", p)
	}
	if p := os.Getenv("https_proxy"); p != "" {
		log.Info("detect https proxy: %s", p)
	}
	go proxy.Serve()

	go proxy.HttpServe(listenPort)

	select {}
}
