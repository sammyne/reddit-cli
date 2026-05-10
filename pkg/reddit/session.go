package reddit

import "sort"

// SessionState holds normalized session information derived from cookies.
type SessionState struct {
	Cookies         map[string]string
	Source          string
	Username        string
	Modhash         string
	LastVerifiedAt  float64
	ValidationError string
	Capabilities    map[string]bool
}

// NewSessionState creates an empty, unauthenticated session.
func NewSessionState() *SessionState {
	return &SessionState{
		Cookies:      make(map[string]string),
		Source:       "none",
		Capabilities: make(map[string]bool),
		ValidationError: "No credential loaded",
	}
}

// SessionFromCredential builds a SessionState from a Credential.
func SessionFromCredential(cred *Credential) *SessionState {
	if cred == nil {
		return NewSessionState()
	}
	cookies := make(map[string]string, len(cred.Cookies))
	for k, v := range cred.Cookies {
		cookies[k] = v
	}
	s := &SessionState{
		Cookies:        cookies,
		Source:         cred.Source,
		Username:       cred.Username,
		Modhash:        cred.Modhash,
		LastVerifiedAt: cred.LastVerifiedAt,
		Capabilities:   make(map[string]bool),
	}
	s.RefreshCapabilities()
	return s
}

// IsAuthenticated returns true if the session has read capability.
func (s *SessionState) IsAuthenticated() bool {
	return s.Capabilities["read"]
}

// CanWrite returns true if the session has write capability.
func (s *SessionState) CanWrite() bool {
	return s.Capabilities["write"]
}

// RefreshCapabilities updates capabilities based on current cookies and modhash.
func (s *SessionState) RefreshCapabilities() {
	caps := make(map[string]bool)
	if s.Cookies["reddit_session"] != "" {
		caps["read"] = true
	}
	modhash := s.Modhash
	if modhash == "" {
		modhash = cookieValue(s.Cookies, "modhash", "csrf_token")
	}
	if modhash != "" {
		s.Modhash = modhash
		caps["write"] = true
	}
	s.Capabilities = caps
}

// ApplyIdentity updates session from a validated /api/me.json response.
func (s *SessionState) ApplyIdentity(identity map[string]interface{}) {
	data := getMap(identity, "data")
	if data == nil {
		data = identity
	}

	if name := AsString(data["name"]); name != "" {
		s.Username = name
	} else if name := AsString(data["username"]); name != "" {
		s.Username = name
	}

	if mh := AsString(data["modhash"]); mh != "" {
		s.Modhash = mh
	} else if s.Modhash == "" {
		s.Modhash = cookieValue(s.Cookies, "modhash", "csrf_token")
	}

	s.ValidationError = ""
	s.RefreshCapabilities()
}

// ApplyValidationError records a validation failure.
func (s *SessionState) ApplyValidationError(message string) {
	s.ValidationError = message
	s.RefreshCapabilities()
}

// SortedCapabilities returns capabilities as a sorted string slice.
func (s *SessionState) SortedCapabilities() []string {
	caps := make([]string, 0, len(s.Capabilities))
	for k, v := range s.Capabilities {
		if v {
			caps = append(caps, k)
		}
	}
	sort.Strings(caps)
	return caps
}

func cookieValue(cookies map[string]string, names ...string) string {
	for _, name := range names {
		if v := cookies[name]; v != "" {
			return v
		}
	}
	return ""
}
