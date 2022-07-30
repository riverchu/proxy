package proxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// open source proxy list
// https://github.com/jhao104/proxy_pool#%E5%85%8D%E8%B4%B9%E4%BB%A3%E7%90%86%E6%BA%90

// proxyHTTPC proxy fetch http client
var proxyHTTPC = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        5,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     10,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		Proxy:               http.ProxyFromEnvironment,
	},
}

// ProxySourceList all proxy source list
var ProxySourceList = [...]Source{localFile, fate0ProxyList, proxy66, mimvp, kxdaili, dieniao}

var _ Source = new(source)

// Source proxy source
type Source interface {
	Name() string
	URL() string
	URLs() []string

	GetProxy() *Proxy
	GetProxies() ProxyArray
	ParseProxy(data io.Reader) ProxyArray

	JudgeQuality() QualityLevel
}

type source struct {
	name     string
	url      string
	urlArray []string

	getProxy   func(*source) *Proxy
	getProxies func(*source) ProxyArray
	parseProxy func(data io.Reader) ProxyArray
}

func (s *source) Name() string   { return s.name }
func (s *source) URL() string    { return s.url }
func (s *source) URLs() []string { return s.urlArray }

func (s *source) GetProxy() *Proxy                     { return s.getProxy(s) }
func (s *source) GetProxies() ProxyArray               { return s.getProxies(s) }
func (s *source) ParseProxy(data io.Reader) ProxyArray { return s.parseProxy(data) }

func (s *source) JudgeQuality() QualityLevel { return s.GetProxies().JudgeQuality() }

var (
	localFile = &source{
		name: "local_file",
		getProxy: func(s *source) (proxy *Proxy) {
			f, err := os.Open("./conf/proxy.json")
			if err != nil {
				return nil
			}
			defer func() { _ = f.Close() }()

			return s.parseProxy(f).Pick()
		},
		getProxies: func(s *source) (proxies ProxyArray) {
			f, err := os.Open("./conf/proxy.json")
			if err != nil {
				return nil
			}
			defer func() { _ = f.Close() }()

			return s.parseProxy(f)
		},
		parseProxy: func(data io.Reader) (proxies ProxyArray) { return nil },
	}

	fate0ProxyList = &source{
		name: "fate0",
		url:  "https://raw.githubusercontent.com/fate0/proxylist/master/proxy.list",
		getProxy: func(s *source) *Proxy {
			resp, err := proxyHTTPC.Get(s.url)
			if err != nil {
				return nil
			}
			defer func() { _ = resp.Body.Close() }()

			return s.parseProxy(resp.Body).Pick()
		},
		getProxies: func(s *source) (proxies ProxyArray) {
			resp, err := proxyHTTPC.Get(s.url)
			if err != nil {
				return nil
			}
			defer func() { _ = resp.Body.Close() }()

			return s.parseProxy(resp.Body)
		},
		parseProxy: func(data io.Reader) (proxies ProxyArray) {
			type proxyItem struct {
				ExportAddress []string `json:"export_address"`
				From          string   `json:"from"`
				Type          string   `json:"type"`
				Country       string   `json:"country"`
				Anonymity     string   `json:"anonymity"`
				ResponseTime  float64  `json:"response_time"`
				Host          string   `json:"host"`
				Port          int      `json:"port"`
			}

			dec := json.NewDecoder(data)
			for {
				var p proxyItem
				if err := dec.Decode(&p); err == io.EOF {
					break
				} else if err != nil {
					_ = err
				}
				if len(p.ExportAddress) == 0 {
					continue
				}
				proxies = append(proxies, &Proxy{
					Scheme:    p.Type,
					Host:      p.ExportAddress[0],
					Port:      p.Port,
					Country:   p.Country,
					Anonymity: p.Anonymity,
					RespTime:  p.ResponseTime,
					Addr:      p.Host,

					Source: "fate0",
				})
			}
			return proxies
		},
	}

	mimvp = &source{
		name: "mimvp-迷扑代理",
		urlArray: []string{
			"https://proxy.mimvp.com/freeopen?proxy=in_hp",
			"https://proxy.mimvp.com/freeopen?proxy=in_tp",
			"https://proxy.mimvp.com/freeopen?proxy=in_socks",
			"https://proxy.mimvp.com/freeopen?proxy=out_hp",
			"https://proxy.mimvp.com/freeopen?proxy=out_tp",
			"https://proxy.mimvp.com/freeopen?proxy=out_socks",
		},
		getProxies: func(s *source) (proxies ProxyArray) {
			for _, u := range s.urlArray {
				resp, err := proxyHTTPC.Get(u)
				if err != nil {
					return nil
				}
				defer func() { _ = resp.Body.Close() }()
				proxies = append(proxies, s.parseProxy(resp.Body)...)
			}
			return proxies
		},
		getProxy: func(s *source) *Proxy {
			if len(s.urlArray) == 0 {
				return nil
			}

			u := s.urlArray[0]
			resp, err := proxyHTTPC.Get(u)
			if err != nil {
				return nil
			}
			defer func() { _ = resp.Body.Close() }()
			return s.parseProxy(resp.Body).Pick()
		},
		parseProxy: func(data io.Reader) (proxies ProxyArray) {
			var reg = regexp.MustCompile(`<td class='free-proxylist-tbl-proxy-ip'>(.*?)</td>` +
				`<td class='free-proxylist-tbl-proxy-port'>.*?port=(.*?) /></td>` +
				`<td class='free-proxylist-tbl-proxy-type' title='\w+'>(.*?)</td>` +
				".*?" +
				`<td class='free-proxylist-tbl-proxy-country'>.*?flags/(.*?)\..*?'.*?</td>`)

			var portMap = map[string]int{"Dgw": 80, "Dgx": 81, "Dgy": 82,
				"DExMA": 110, "DgwOA": 808, "Dk5OQ": 999, "DEwODA": 1080, "DQxNDU": 4145, "DQxNTM": 4153,
				"DMxMjg": 3128, "DgwMDA": 8000, "DgwMDE": 8001, "DgwODA": 8080, "DgwODE": 8081, "Dg4ODg": 8888, "Dk3OTc": 9797, "Dk5OTk": 9999,
				"DEwODAx": 10801, "DQyMDU1": 42055, "DUzMjgx": 53281, "DU1NDQz": 55443}

			content, err := ioutil.ReadAll(data)
			if err != nil {
				return
			}
			if !bytes.Contains(content, []byte("<tbody>")) || !bytes.Contains(content, []byte("</tbody>")) {
				return
			}
			content = content[bytes.Index(content, []byte("<tbody>"))+len("<tbody>") : bytes.Index(content, []byte("</tbody>"))]
			items := strings.Split(string(bytes.TrimSpace(content)), "<tr>")
			for _, i := range items {
				for _, info := range reg.FindAllStringSubmatch(strings.TrimSuffix(i, "</tr>"), -1) {
					ip, port, typ, country := info[1], info[2], info[3], info[4]
					port = strings.ReplaceAll(port[14:], "O0O", "")

					proxies = append(proxies, &Proxy{
						Scheme:  strings.ToLower(strings.Split(typ, "/")[0]),
						Host:    ip,
						Port:    portMap[port],
						Country: strings.ToUpper(country),

						Source: "mimvp",
					})
				}
			}
			return proxies
		},
	}

	proxy66 = &source{
		// http://www.66ip.cn/
		name:     "66代理",
		urlArray: []string{"http://www.66ip.cn/mo.php"},
		getProxies: func(s *source) (proxies ProxyArray) {
			for _, u := range s.urlArray {
				resp, err := proxyHTTPC.Get(u)
				if err != nil {
					return nil
				}
				defer func() { _ = resp.Body.Close() }()
				proxies = append(proxies, s.parseProxy(resp.Body)...)
			}
			return proxies
		},
		getProxy: func(s *source) (proxy *Proxy) {
			if len(s.urlArray) == 0 {
				return nil
			}

			u := s.urlArray[0]
			resp, err := proxyHTTPC.Get(u)
			if err != nil {
				return nil
			}
			defer func() { _ = resp.Body.Close() }()
			return s.parseProxy(resp.Body).Pick()
		},
		parseProxy: func(data io.Reader) (proxies ProxyArray) {
			reg := regexp.MustCompile(`((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}:\d+`)

			content, err := ioutil.ReadAll(data)
			if err != nil {
				return nil
			}
			for _, item := range reg.FindAllString(string(content), -1) {
				info := strings.Split(item, ":")
				if len(info) != 2 {
					continue
				}
				port, err := strconv.Atoi(info[1])
				if err != nil {
					continue
				}
				proxies = append(proxies, &Proxy{
					Scheme: "http",
					Host:   info[0],
					Port:   port,

					Source: "proxy66",
				})
			}
			return
		},
	}

	kxdaili = &source{
		// http://www.kxdaili.com/
		name:     "开心代理",
		urlArray: []string{"http://www.kxdaili.com/dailiip.html", "http://www.kxdaili.com/dailiip/2/1.html"},
		getProxies: func(s *source) (proxies ProxyArray) {
			for _, u := range s.urlArray {
				resp, err := proxyHTTPC.Get(u)
				if err != nil {
					return nil
				}
				defer func() { _ = resp.Body.Close() }()
				proxies = append(proxies, s.parseProxy(resp.Body)...)
			}
			return proxies
		},
		getProxy: func(s *source) (proxy *Proxy) {
			if len(s.urlArray) == 0 {
				return nil
			}

			u := s.urlArray[0]
			resp, err := proxyHTTPC.Get(u)
			if err != nil {
				return nil
			}
			defer func() { _ = resp.Body.Close() }()
			return s.parseProxy(resp.Body).Pick()
		},
		parseProxy: func(data io.Reader) (proxies ProxyArray) {
			reg := regexp.MustCompile(`<td>([\d.]+)</td>\s+<td>(\d+)</td>\s+.*?</td>\s+<td>([\w,]+)</td>`)

			content, err := ioutil.ReadAll(data)
			if err != nil {
				return nil
			}
			for _, info := range reg.FindAllStringSubmatch(string(content), -1) {
				ip, portStr, scheme := info[1], info[2], info[3]
				port, err := strconv.Atoi(portStr)
				if err != nil {
					continue
				}
				proxies = append(proxies, &Proxy{
					Scheme: strings.ToLower(strings.Split(scheme, ",")[0]),
					Host:   ip,
					Port:   port,

					Source: "kxdaili",
				})
			}
			return
		},
	}

	dieniao = &source{
		// https://www.dieniao.com/
		name: "蝶鸟",
		url:  "https://www.dieniao.com/FreeProxy.html",
		getProxies: func(s *source) (proxies ProxyArray) {
			resp, err := proxyHTTPC.Get(s.url)
			if err != nil {
				return nil
			}
			defer func() { _ = resp.Body.Close() }()
			return s.parseProxy(resp.Body)
		},
		getProxy: func(s *source) (proxy *Proxy) {
			resp, err := proxyHTTPC.Get(s.url)
			if err != nil {
				return nil
			}
			defer func() { _ = resp.Body.Close() }()
			return s.parseProxy(resp.Body).Pick()
		},
		parseProxy: func(data io.Reader) (proxies ProxyArray) {
			reg := regexp.MustCompile(`<span class='f-address'>([\d.]+)</span>\s+<span class='f-port'>(\d+)</span>`)

			content, err := ioutil.ReadAll(data)
			if err != nil {
				return nil
			}
			for _, info := range reg.FindAllStringSubmatch(string(content), -1) {
				ip, portStr := info[1], info[2]
				port, err := strconv.Atoi(portStr)
				if err != nil {
					continue
				}
				proxies = append(proxies, &Proxy{
					Scheme: "http",
					Host:   ip,
					Port:   port,

					Source: "dieniao",
				})
			}
			return
		},
	}
)
