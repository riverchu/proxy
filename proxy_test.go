package proxy

import (
	"testing"
	"time"
)

func TestGetProxy(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Logf("Got Proxy: %s", GetProxy().String())
	}
}

func Test_JudgedProxy(t *testing.T) {
	t.Logf("fate0 quality: %v", fate0ProxyList.JudgeQuality())
}

func Test_ICMP(t *testing.T) {
	delay, err := (&Proxy{}).ICMPTest("127.0.0.1")
	if err != nil {
		t.Errorf("Failed to send ICMP: %v", err)
	}
	t.Logf("target delay: %s", delay)
}

func Test_Get(t *testing.T) {
	delay, err := (&Proxy{Scheme: "http", Host: "127.0.0.1", Port: 1080}).GETTest()
	if err != nil {
		t.Errorf("Fail to Get: %s", err)
	}
	t.Logf("target delay: %s", delay)
}

func Test_FilterProxy(t *testing.T) {
	defaultServer.JudgeQuality()
	proxies := defaultServer.Filter(FilterLowProxy).GetProxies()
	for _, p := range proxies {
		t.Logf("get proxy: %s", p)
	}
}

func TestProxyServe(t *testing.T) {
	t.Log("start serve")

	go serve()
	for range time.Tick(time.Second) {
		t.Logf("%s got %d proxies\n", time.Now(), len(GetProxies()))
	}
}

func TestProxyServe_Renew(t *testing.T) {
	(&Server{sources: ProxySourceList[:]}).Renew()
}

func TestSource_GetProxies(t *testing.T) {
	t.Logf("mimvp get proxies: %d", len(dieniao.GetProxies()))
	t.Logf("proxy66 get proxies: %d", len(dieniao.GetProxies()))
	t.Logf("kxdaili get proxies: %d", len(dieniao.GetProxies()))
	t.Logf("dieniao get proxies: %d", len(dieniao.GetProxies()))
}
