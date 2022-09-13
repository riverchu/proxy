package proxy

// FilterOption ...
type FilterOption func(*Proxy) (pass bool)

var (
	// FilterProxyLevel filter low quality
	FilterProxyLevel = func(level QualityLevel) FilterOption {
		return FilterProxy(level.Threshold())
	}

	// FilterProxy filter proxy with quality
	FilterProxy = func(quality Quality) FilterOption {
		return func(p *Proxy) bool { return p.quality >= quality }
	}

	// FilterN filter n proxies
	FilterN = func(n int) FilterOption {
		if n <= 0 {
			return func(*Proxy) bool { return false }
		}

		return func(*Proxy) bool {
			defer func() { n-- }()
			return n > 0
		}
	}
)
