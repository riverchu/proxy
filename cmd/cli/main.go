package main

import (
	"os"
	"time"

	"github.com/riverchu/pkg/log"
	"github.com/riverchu/proxy"
	_ "github.com/riverchu/proxy/source"
)

const interval = 10 * time.Second

func main() {
	log.Info("this is a proxy provider")
	if p := os.Getenv("http_proxy"); p != "" {
		log.Info("detect http proxy: %s", p)
	}
	if p := os.Getenv("https_proxy"); p != "" {
		log.Info("detect https proxy: %s", p)
	}

	go proxy.Serve()

	log.Info("refresh interval: %s", interval)
	for range time.Tick(interval) {
		log.Info("proxy refreshing")
		for _, p := range proxy.GetProxies(proxy.FilterProxyLevel(proxy.MEDIUM)).String() {
			log.Info("got proxy: %s", p)
		}
	}
}
