package proxy

// Quality 代理质量，越高越好
type Quality int64

// Judge ...
func (q Quality) Judge() QualityLevel {
	switch {
	case q >= HIGH.Threshold():
		return HIGH
	case q >= MEDIUM.Threshold():
		return MEDIUM
	case q > LOW.Threshold():
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

func (p QualityLevel) Threshold() Quality {
	switch p {
	case HIGH:
		return 100
	case MEDIUM:
		return 50
	case LOW:
		return 20
	case UNAVAILABLE:
		return 0
	default:
		return 0
	}
}
