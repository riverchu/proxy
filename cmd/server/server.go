package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/riverchu/pkg/log"
)

const listenPort = "8080"

func main() {
	l, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Error("listening port %d fail: %s", err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Error("accept connection fail: %s", err)
		}
		go handleClientRequest(client)
	}
}

// https://www.flysnow.org/2016/12/24/golang-http-proxy
func handleClientRequest(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		log.Info("read fail: %s", err)
		return
	}

	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)

	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Info("read fail: %s", err)
		return
	}

	if hostPortURL.Opaque == "443" { //https访问
		address = hostPortURL.Scheme + ":443"
	} else { //http访问
		if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	}

	//获得了请求的host和port，就开始拨号吧
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Info("dail fail: %s", err)
		return
	}
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else {
		server.Write(b[:n])
	}
	//进行转发
	go io.Copy(server, client)
	io.Copy(client, server)
}
