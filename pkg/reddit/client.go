package reddit

import (
	"errors"
	"fmt"
	"strconv"
)

// RedditClient provides high-level access to Reddit's JSON API.
type RedditClient struct {
	Session   *SessionState
	transport *ReadTransport
}

// NewRedditClient creates a client from a credential.
func NewRedditClient(cred *Credential) *RedditClient {
	session := SessionFromCredential(cred)
	return &RedditClient{Session: session}
}

// Open initializes the HTTP transport. Must be called before making requests.
func (c *RedditClient) Open() {
	fp := NewBrowserFingerprint()
	cfg := DefaultConfig()
	c.transport = NewReadTransport(c.Session, cfg, fp)
}

// Close releases transport resources.
func (c *RedditClient) Close() {
	if c.transport != nil {
		c.transport.Close()
		c.transport = nil
	}
}

func (c *RedditClient) get(path string, params map[string]string) (map[string]interface{}, error) {
	if c.transport == nil {
		return nil, &RedditAPIError{Message: "Client not initialized. Call Open() first."}
	}
	return c.transport.Request("GET", path, params)
}

// GetMe fetches the current user's identity and enriches the session.
func (c *RedditClient) GetMe() (map[string]interface{}, error) {
	data, err := c.get(MeURL, map[string]string{"raw_json": "1"})
	if err != nil {
		return nil, err
	}
	c.Session.ApplyIdentity(data)
	return data, nil
}

// ValidateSession probes the current credential and returns status info.
func (c *RedditClient) ValidateSession() map[string]interface{} {
	identity, err := c.GetMe()
	if err != nil {
		var apiErr *RedditAPIError
		if errors.As(err, &apiErr) {
			c.Session.ApplyValidationError(apiErr.Message)
		}
		return map[string]interface{}{
			"authenticated":  false,
			"username":       c.Session.Username,
			"capabilities":   c.Session.SortedCapabilities(),
			"modhash_present": c.Session.Modhash != "",
			"error":          err.Error(),
		}
	}
	return map[string]interface{}{
		"authenticated":  true,
		"username":       c.Session.Username,
		"capabilities":   c.Session.SortedCapabilities(),
		"modhash_present": c.Session.Modhash != "",
		"identity":       identity,
	}
}

// GetUserAbout fetches a user's profile info.
func (c *RedditClient) GetUserAbout(username string) (map[string]interface{}, error) {
	path := fmt.Sprintf(UserAboutURL, username)
	data, err := c.get(path, map[string]string{"raw_json": "1"})
	if err != nil {
		return nil, err
	}
	if inner := getMap(data, "data"); inner != nil {
		return inner, nil
	}
	return data, nil
}

// Search searches Reddit posts.
func (c *RedditClient) Search(query, subreddit, sort, timeFilter string, limit int, after string) (map[string]interface{}, error) {
	var path string
	if subreddit != "" {
		path = fmt.Sprintf(SubredditSearchURL, subreddit)
	} else {
		path = SearchURL
	}

	params := map[string]string{
		"q":           query,
		"sort":        sort,
		"t":           timeFilter,
		"limit":       strconv.Itoa(limit),
		"restrict_sr": "off",
		"raw_json":    "1",
	}
	if subreddit != "" {
		params["restrict_sr"] = "on"
	}
	if after != "" {
		params["after"] = after
	}
	return c.get(path, params)
}
