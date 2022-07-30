package proxy

// FilterOption ...
type FilterOption func(*Proxy) (pass bool)

var (
	// FilterLowProxy filter low quality
	FilterLowProxy FilterOption = func(p *Proxy) bool {
		return p.QualityLevel() >= MEDIUM
	}

	// FilterProxyQuality filter proxy with quality
	FilterProxyQuality = func(quality Quality) FilterOption {
		return func(p *Proxy) bool { return p.quality >= quality }
	}

	// FilterN filter n proxies
	FilterN = func(n int) FilterOption {
		if n <= 0 {
			return func(*Proxy) bool { return false }
		}

		count := 0
		return func(p *Proxy) bool {
			defer func() { count++ }()
			return count < n
		}
	}
)
