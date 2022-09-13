package proxy

import (
	"io"
)

// Source proxy source
type Source interface {
	Name() string // return unique name
	URL() string
	URLs() []string

	GetProxy() *Proxy
	GetProxies() ProxyArray
	ParseProxy(data io.Reader) ProxyArray

	JudgeQuality() QualityLevel
}
