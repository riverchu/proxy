package main

import (
	"fmt"
	"os"
	"time"

	"github.com/riverchu/proxy"
)

func main() {
	fmt.Println("this is a proxy server")
	fmt.Println("http proxy", os.Getenv("http_proxy"))
	fmt.Println("https proxy", os.Getenv("https_proxy"))

	go proxy.Serve()

	for range time.Tick(10 * time.Second) {
		fmt.Println("fetch result:")
		for _, p := range proxy.GetProxies().String() {
			fmt.Println("proxy:", p)
		}
	}
}
