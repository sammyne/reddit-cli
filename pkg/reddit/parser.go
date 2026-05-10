package reddit

import (
	"encoding/json"
	"fmt"
)

// asInt safely converts an interface{} value to int.
func asInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	case nil:
		return 0
	default:
		return 0
	}
}

// asFloat safely converts an interface{} value to float64.
func asFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	case nil:
		return 0
	default:
		return 0
	}
}

// AsString safely converts an interface{} value to string.
func AsString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// asBool safely converts an interface{} value to bool.
func asBool(v interface{}, defaultVal bool) bool {
	if v == nil {
		return defaultVal
	}
	b, ok := v.(bool)
	if ok {
		return b
	}
	return defaultVal
}

// getMap safely gets a nested map from a map.
func getMap(m map[string]interface{}, key string) map[string]interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	mm, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return mm
}

// getSlice safely gets a slice from a map.
func getSlice(m map[string]interface{}, key string) []interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	s, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return s
}

// ParsePost parses a Reddit post payload into a Post struct.
func ParsePost(payload map[string]interface{}) Post {
	return Post{
		ID:          AsString(payload["id"]),
		Name:        AsString(payload["name"]),
		Title:       AsString(payload["title"]),
		Subreddit:   AsString(payload["subreddit"]),
		Author:      AsString(payload["author"]),
		Score:       asInt(payload["score"]),
		NumComments: asInt(payload["num_comments"]),
		CreatedUTC:  asFloat(payload["created_utc"]),
		Permalink:   AsString(payload["permalink"]),
		URL:         AsString(payload["url"]),
		Selftext:    AsString(payload["selftext"]),
		IsSelf:      asBool(payload["is_self"], true),
		Over18:      asBool(payload["over_18"], false),
		IsVideo:     asBool(payload["is_video"], false),
		Stickied:    asBool(payload["stickied"], false),
	}
}

// ParseListing parses a Reddit listing JSON response into a ListingPage.
func ParseListing(data map[string]interface{}) ListingPage {
	listing := getMap(data, "data")
	if listing == nil {
		return ListingPage{}
	}

	children := getSlice(listing, "children")
	posts := make([]Post, 0, len(children))
	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}
		postData := getMap(childMap, "data")
		if postData == nil {
			postData = childMap
		}
		posts = append(posts, ParsePost(postData))
	}

	page := ListingPage{Items: posts}
	if after, ok := listing["after"].(string); ok && after != "" {
		page.After = &after
	}
	if before, ok := listing["before"].(string); ok && before != "" {
		page.Before = &before
	}
	return page
}

// ParseUserProfile parses a user profile API response into a UserProfile.
func ParseUserProfile(data map[string]interface{}) UserProfile {
	inner := getMap(data, "data")
	if inner == nil {
		inner = data
	}
	return UserProfile{
		Name:         AsString(inner["name"]),
		LinkKarma:    asInt(inner["link_karma"]),
		CommentKarma: asInt(inner["comment_karma"]),
		CreatedUTC:   asFloat(inner["created_utc"]),
		IsGold:       asBool(inner["is_gold"], false),
		IsMod:        asBool(inner["is_mod"], false),
	}
}
