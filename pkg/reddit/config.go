package reddit

import "time"

// RuntimeConfig holds transport and anti-detection defaults.
type RuntimeConfig struct {
	Timeout           time.Duration
	ReadRequestDelay  time.Duration
	WriteRequestDelay time.Duration
	MaxRetries        int
	StatusCheckTimeout time.Duration
}

// DefaultConfig returns the default runtime configuration.
func DefaultConfig() RuntimeConfig {
	return RuntimeConfig{
		Timeout:            30 * time.Second,
		ReadRequestDelay:   1 * time.Second,
		WriteRequestDelay:  2500 * time.Millisecond,
		MaxRetries:         3,
		StatusCheckTimeout: 10 * time.Second,
	}
}
