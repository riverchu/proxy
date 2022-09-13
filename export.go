package proxy

import (
	"github.com/riverchu/pkg/log"
)

func init() {
	loadFromDB()
	loadFromFlie()
}

// Serve singlton Serve
func Serve(sources ...Source) {
	log.Info("Proxy Server Starting...")
	defer log.Info("Proxy Server Stopped...")

	serve(sources...)
}

// GetProxy get one proxy
func GetProxy(opts ...FilterOption) *Proxy {
	return defaultServer.GetProxy(opts...)
}

// GetProxies get all proxies
func GetProxies(opts ...FilterOption) ProxyArray {
	return defaultServer.GetProxies(opts...)
}

// RegisterSource register source
func RegisterSource(sources ...Source) {
	defaultServer.RegisterSource(sources...)
}
