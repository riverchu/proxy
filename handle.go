package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/riverchu/pkg/log"
)

// DirectProxyConn proxy request
// https://www.flysnow.org/2016/12/24/golang-http-proxy
func DirectProxyConn(client net.Conn) {
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

	if hostPortURL.Opaque == "443" { // https访问
		address = hostPortURL.Scheme + ":443"
	} else { // http访问
		if strings.Index(hostPortURL.Host, ":") < 0 { // 默认追加80端口
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

// ProxyConn proxy connection
func ProxyConn(client net.Conn) {
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

	var method, host string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)

	//获得了请求的host和port，就开始拨号吧
	proxyAddr := GetProxy().Target()
	log.Info("using proxy: %s", proxyAddr)

	server, err := net.Dial("tcp", proxyAddr)
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
