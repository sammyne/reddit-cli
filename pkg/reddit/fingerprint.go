package reddit

import (
	"fmt"
	"runtime"
)

// BrowserFingerprint holds consistent request identity across transports.
type BrowserFingerprint struct {
	UserAgent      string
	SecChUA        string
	SecChUAMobile  string
	SecChUAPlatform string
	AcceptLanguage string
}

// NewBrowserFingerprint creates a fingerprint based on the current OS.
func NewBrowserFingerprint() BrowserFingerprint {
	var osPart, platform string
	switch runtime.GOOS {
	case "darwin":
		osPart = "Macintosh; Intel Mac OS X 10_15_7"
		platform = `"macOS"`
	case "windows":
		osPart = "Windows NT 10.0; Win64; x64"
		platform = `"Windows"`
	default: // linux and others
		osPart = "X11; Linux x86_64"
		platform = `"Linux"`
	}

	ua := fmt.Sprintf(
		"Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		osPart,
	)

	return BrowserFingerprint{
		UserAgent:       ua,
		SecChUA:         `"Chromium";v="133", "Not(A:Brand";v="99", "Google Chrome";v="133"`,
		SecChUAMobile:   "?0",
		SecChUAPlatform: platform,
		AcceptLanguage:  "en-US,en;q=0.9",
	}
}

// ReadHeaders returns headers for read (GET) requests.
func (f BrowserFingerprint) ReadHeaders() map[string]string {
	return map[string]string{
		"User-Agent":       f.UserAgent,
		"sec-ch-ua":        f.SecChUA,
		"sec-ch-ua-mobile": f.SecChUAMobile,
		"sec-ch-ua-platform": f.SecChUAPlatform,
		"Sec-Fetch-Dest":   "empty",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Site":   "same-origin",
		"Accept":           "application/json, text/plain, */*",
		"Accept-Language":  f.AcceptLanguage,
	}
}
