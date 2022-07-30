package proxy

// Quality 代理质量，越高越好
type Quality int64

// Judge ...
func (q Quality) Judge() QualityLevel {
	switch {
	case q >= 100:
		return HIGH
	case q >= 50:
		return MEDIUM
	case q > 20:
		return LOW
	default:
		return UNAVAILABLE
	}
}

// QualityLevel ...
type QualityLevel int64

const (
	// UNAVAILABLE ...
	UNAVAILABLE QualityLevel = iota
	// LOW ...
	LOW
	// MEDIUM ...
	MEDIUM
	// HIGH ...
	HIGH
)

func (p QualityLevel) String() string {
	switch p {
	case UNAVAILABLE:
		return "UNAVAILABLE"
	case LOW:
		return "LOW"
	case MEDIUM:
		return "MEDIUM"
	case HIGH:
		return "HIGH"
	default:
		return "UNKNOWN"
	}
}
