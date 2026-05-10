package reddit

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type indexCachePayload struct {
	Source  string                   `json:"source"`
	SavedAt float64                 `json:"saved_at"`
	Count   int                     `json:"count"`
	Items   []map[string]interface{} `json:"items"`
}

// SaveIndex saves a list of posts to the index cache file.
func SaveIndex(posts []Post, source string) {
	if len(posts) == 0 {
		return
	}
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Printf("failed to create config dir: %v", err)
		return
	}

	items := make([]map[string]interface{}, 0, len(posts))
	for _, p := range posts {
		entry := map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"title":        p.Title,
			"subreddit":    p.Subreddit,
			"author":       p.Author,
			"score":        p.Score,
			"num_comments": p.NumComments,
			"permalink":    p.Permalink,
			"url":          p.URL,
		}
		if p.ID != "" {
			items = append(items, entry)
		}
	}

	payload := indexCachePayload{
		Source:  source,
		SavedAt: float64(time.Now().Unix()),
		Count:   len(items),
		Items:   items,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("failed to marshal index cache: %v", err)
		return
	}
	if err := os.WriteFile(IndexCacheFile(), data, 0o600); err != nil {
		log.Printf("failed to write index cache: %v", err)
	}
}

// GetItemByIndex returns a cached item by 1-based index, or nil.
func GetItemByIndex(index int) map[string]interface{} {
	if index <= 0 {
		return nil
	}
	data, err := os.ReadFile(IndexCacheFile())
	if err != nil {
		return nil
	}
	var payload indexCachePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}
	if index > len(payload.Items) {
		return nil
	}
	return payload.Items[index-1]
}

// IndexInfo holds metadata about the current index cache.
type IndexInfo struct {
	Exists  bool    `json:"exists"`
	Count   int     `json:"count"`
	Source  string  `json:"source,omitempty"`
	SavedAt float64 `json:"saved_at,omitempty"`
}

// GetIndexInfo returns metadata about the current index cache.
func GetIndexInfo() IndexInfo {
	data, err := os.ReadFile(IndexCacheFile())
	if err != nil {
		return IndexInfo{Exists: false}
	}
	var payload indexCachePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return IndexInfo{Exists: false}
	}
	return IndexInfo{
		Exists:  true,
		Count:   payload.Count,
		Source:  payload.Source,
		SavedAt: payload.SavedAt,
	}
}
