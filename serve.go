package proxy

import (
	"net"

	"github.com/riverchu/pkg/log"
)

func HttpServe(port string) {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Error("listening port %d fail: %s", err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Error("accept connection fail: %s", err)
		}
		go ProxyConn(client)
	}
}
