package proxy

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/riverchu/pkg/log"
	"github.com/riverchu/pkg/netool"
	"github.com/riverchu/pkg/thread"
)

// Proxy ...
type Proxy struct {
	Scheme string // 协议
	Host   string // 地址
	Port   int    // 端口

	Source    string
	Type      string
	Country   string
	Anonymity string
	Addr      string
	RespTime  float64
	Ping      float64

	mu           sync.RWMutex
	quality      Quality      // 质量分
	qualityLevel QualityLevel // 质量水平
}

// AccessQuality ...
func (p *Proxy) AccessQuality() (quality Quality) {
	defer func() {
		p.mu.Lock()
		p.quality = quality
		p.mu.Unlock()
	}()

	if !p.isValid() {
		return 0
	}

	// p.accessByICMP()
	return p.accessByGET()
}

// AccessQualityLevel 评估质量级别
func (p *Proxy) AccessQualityLevel() QualityLevel {
	level := p.AccessQuality().Judge()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.qualityLevel = level

	return level
}

// func (p *Proxy) accessByICMP() (quality ProxyQuality) {
// 	if delay, err := p.ICMPTest(p.Host); err != nil {
// 		log.Warn("Proxy %q icmp test fail: %s", p.GetProxy(), err)
// 	} else if delay > time.Second { // 0 point
// 	} else {
// 		quality += ProxyQuality((1000 - delay/time.Millisecond) / 10)
// 	}
// 	return
// }

func (p *Proxy) accessByGET() (quality Quality) {
	if delay, err := p.GETTest(); err != nil {
		log.Warn("Proxy %q get test fail: %s", p.String(), err)
	} else if delay > time.Second { // 0 point scores
	} else {
		quality += Quality((1000 - delay/time.Millisecond) / 10)
	}
	return quality
}

func (p *Proxy) isValid() bool {
	// net.ParseIP(p.Host) == nil
	return p.Port != 0 && (p.Scheme == "http" || p.Scheme == "https" || p.Scheme == "socks4" || p.Scheme == "socks5")
}

// Quality ...
func (p *Proxy) Quality() Quality { return p.quality }

// QualityLevel ...
func (p *Proxy) QualityLevel() QualityLevel { return p.qualityLevel }

// String return proxy url as string
func (p *Proxy) String() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%s://%s:%d", p.Scheme, p.Host, p.Port)
}

// Target return host with port
func (p *Proxy) Target() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

// URL convert to url
func (p *Proxy) URL() *url.URL {
	if p == nil {
		return nil
	}

	u, err := url.Parse(p.String())
	if err != nil {
		log.Warn("Proxy %q parse fail: %s", p.String(), err)
		return nil
	}
	return u
}

// ICMPTest ...
func (p *Proxy) ICMPTest(host string) (time.Duration, error) {
	if net.ParseIP(host) == nil {
		ips, err := p.lookupIP(host)
		if err != nil {
			return 0, fmt.Errorf("lookup ip fail: %s", err)
		}
		host = ips[0].String()
	}
	return netool.ICMPDelay(host, 6)
}

var reqHost = [...]string{
	"http://qq.com",
	// "http://www.baidu.com",
	// "http://google.com",
}

// GETTest ...
func (p *Proxy) GETTest() (time.Duration, error) {
	var client = &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(p.URL()),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var sum time.Duration
	for _, host := range reqHost {
		start := time.Now()
		resp, err := client.Get(host)
		if err != nil {
			sum += 2 * time.Second
			continue
		}
		defer resp.Body.Close() // nolint
		sum += time.Since(start)
		log.Debug("Proxy(%s) request %s cost: %s", p.String(), host, time.Since(start))
	}
	return sum / time.Duration(len(reqHost)), nil
}

func (p *Proxy) lookupIP(domain string) ([]net.IP, error) {
	ips, err := netool.LookupIP(domain)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("ip not found")
	}
	return ips, nil
}

// ProxyArray proxy array
type ProxyArray []*Proxy // nolint

func (a ProxyArray) Len() int           { return len(a) }
func (a ProxyArray) Less(i, j int) bool { return a[i].quality < a[j].quality }
func (a ProxyArray) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// String convert to string
func (a ProxyArray) String() (proxies []string) {
	for _, p := range a {
		proxies = append(proxies, p.String())
	}
	return proxies
}

// URL convert to URL
func (a ProxyArray) URL() (proxies []*url.URL) {
	for _, p := range a {
		proxies = append(proxies, p.URL())
	}
	return proxies
}

// Pick pick one proxy
func (a ProxyArray) Pick() *Proxy {
	if len(a) == 0 {
		return nil
	}
	return a[rand.Intn(len(a))]
}

// JudgeQuality judge proxy quality
func (a ProxyArray) JudgeQuality() QualityLevel {
	var mu sync.Mutex
	var count int

	pool := thread.NewTimeoutPoolWithDefaults()
	for _, proxy := range a {
		pool.Submit(&thread.Job{
			Handler: func(v ...interface{}) {
				p := v[0].(*Proxy)
				if p.AccessQualityLevel() == HIGH {
					mu.Lock()
					count++
					mu.Unlock()
				}
			},
			Params: []interface{}{proxy},
		})
	}
	pool.StartAndWait(300 * time.Millisecond * time.Duration(len(a)))

	ratio := float64(count) / float64(len(a))
	switch {
	case ratio > 0.8:
		return HIGH
	case ratio > 0.5:
		return MEDIUM
	default:
		return LOW
	}
}
