package source

import "testing"

func Test_JudgedProxy(t *testing.T) {
	t.Logf("fate0 quality: %v", fate0ProxyList.JudgeQuality())
}

func TestSource_GetProxies(t *testing.T) {
	t.Logf("mimvp get proxies: %d", len(mimvp.GetProxies()))
	t.Logf("proxy66 get proxies: %d", len(proxy66.GetProxies()))
	t.Logf("kxdaili get proxies: %d", len(kxdaili.GetProxies()))
	t.Logf("dieniao get proxies: %d", len(dieniao.GetProxies()))
}
