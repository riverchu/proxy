package proxy

import (
	"fmt"
	"net"

	"github.com/riverchu/pkg/log"
)

func HttpServe(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Error("listening port %d fail: %s", port, err)
		return
	}
	log.Info("listening port %d", port)

	for {
		client, err := l.Accept()
		if err != nil {
			log.Error("accept connection fail: %s", err)
		}
		go ProxyConn(client)
	}
}
