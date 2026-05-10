package reddit

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/chrome"
	_ "github.com/browserutils/kooky/browser/edge"
	_ "github.com/browserutils/kooky/browser/firefox"
	_ "github.com/browserutils/kooky/browser/brave"
)

const credentialTTLDays = 7
const credentialTTLSeconds = credentialTTLDays * 86400

// Credential holds Reddit session cookies.
type Credential struct {
	Cookies        map[string]string `json:"cookies"`
	Source         string            `json:"source"`
	Username       string            `json:"username,omitempty"`
	Modhash        string            `json:"modhash,omitempty"`
	SavedAt        float64           `json:"saved_at,omitempty"`
	LastVerifiedAt float64           `json:"last_verified_at,omitempty"`
}

// IsValid returns true if the credential has cookies.
func (c *Credential) IsValid() bool {
	return c != nil && len(c.Cookies) > 0
}

// CookieHeader returns cookies formatted as a Cookie header value.
func (c *Credential) CookieHeader() string {
	if c == nil {
		return ""
	}
	var s string
	for k, v := range c.Cookies {
		if s != "" {
			s += "; "
		}
		s += k + "=" + v
	}
	return s
}

// SaveCredential writes the credential to disk with 0600 permissions.
func SaveCredential(cred *Credential) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if cred.SavedAt == 0 {
		cred.SavedAt = float64(time.Now().Unix())
	}
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(CredentialFile(), data, 0o600)
}

// LoadCredential loads a saved credential with TTL-based auto-refresh.
func LoadCredential() *Credential {
	data, err := os.ReadFile(CredentialFile())
	if err != nil {
		return nil
	}
	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil
	}
	if !cred.IsValid() {
		return nil
	}
	// TTL check
	if cred.SavedAt > 0 && (float64(time.Now().Unix())-cred.SavedAt) > float64(credentialTTLSeconds) {
		log.Printf("Credential older than %d days, attempting browser refresh", credentialTTLDays)
		if fresh := ExtractBrowserCredential(); fresh != nil {
			log.Printf("Auto-refreshed credential from browser")
			return fresh
		}
		log.Printf("Cookie refresh failed; using existing cookies")
	}
	return &cred
}

// ClearCredential removes the saved credential file.
func ClearCredential() {
	_ = os.Remove(CredentialFile())
}

// ExtractBrowserCredential extracts Reddit cookies from installed browsers.
func ExtractBrowserCredential() *Credential {
	ctx := context.Background()
	cookies := make(map[string]string)

	for c, err := range kooky.TraverseCookies(ctx, kooky.DomainHasSuffix(".reddit.com"), kooky.Valid) {
		if err != nil {
			continue
		}
		cookies[c.Name] = c.Value
	}

	if len(cookies) == 0 {
		return nil
	}
	// Check for required cookie
	hasRequired := false
	for _, name := range RequiredCookies {
		if cookies[name] != "" {
			hasRequired = true
			break
		}
	}
	if !hasRequired {
		return nil
	}
	cred := &Credential{
		Cookies: cookies,
		Source:  "browser",
	}
	if err := SaveCredential(cred); err != nil {
		log.Printf("failed to save credential: %v", err)
	}
	return cred
}

// GetCredential tries saved → browser → nil.
func GetCredential() *Credential {
	if cred := LoadCredential(); cred != nil {
		return cred
	}
	return ExtractBrowserCredential()
}
