// Package reddit provides the core library for the Reddit CLI client.
package reddit

import (
	"os"
	"path/filepath"
)

// Base URL for Reddit's public JSON API.
const BaseURL = "https://www.reddit.com"

// API endpoint paths.
const (
	HomeURL             = "/.json"
	PopularURL          = "/r/popular.json"
	AllURL              = "/r/all.json"
	SearchURL           = "/search.json"
	SubredditSearchURL  = "/r/%s/search.json"   // %s = subreddit
	SubredditAboutURL   = "/r/%s/about.json"    // %s = subreddit
	UserAboutURL        = "/user/%s/about.json" // %s = username
	MeURL               = "/api/me.json"
	MorechildrenURL     = "/api/morechildren.json"
	PostCommentsURL     = "/r/%s/comments/%s.json"  // %s = subreddit, post_id
	PostCommentsShortURL = "/comments/%s.json"       // %s = post_id
)

// Required cookies for authenticated sessions.
var RequiredCookies = []string{"reddit_session"}

// Sort options for subreddit listings.
var SortOptions = []string{"hot", "new", "top", "rising", "controversial", "best"}

// Time filter options for top/controversial.
var TimeFilters = []string{"hour", "day", "week", "month", "year", "all"}

// Search sort options.
var SearchSortOptions = []string{"relevance", "hot", "top", "new", "comments"}

// Default and maximum page sizes.
const (
	DefaultLimit = 25
	MaxLimit     = 100
)

// configDirPrimary is the preferred config directory name.
const configDirPrimary = "reddit"

// configDirFallback is used when the primary directory does not exist.
const configDirFallback = "rdt-cli"

// ConfigDir returns the configuration directory path.
// It prefers $HOME/.config/reddit/ and falls back to $HOME/.config/rdt-cli/.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	primary := filepath.Join(home, ".config", configDirPrimary)
	if info, err := os.Stat(primary); err == nil && info.IsDir() {
		return primary
	}
	fallback := filepath.Join(home, ".config", configDirFallback)
	if info, err := os.Stat(fallback); err == nil && info.IsDir() {
		return fallback
	}
	// Neither exists; return primary so it will be created on first write.
	return primary
}

// CredentialFile returns the path to the credential JSON file.
func CredentialFile() string {
	return filepath.Join(ConfigDir(), "credential.json")
}

// IndexCacheFile returns the path to the index cache JSON file.
func IndexCacheFile() string {
	return filepath.Join(ConfigDir(), "index_cache.json")
}
