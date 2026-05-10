package reddit

import (
	"reflect"
	"testing"
)

func TestParsePost(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]interface{}
		want    Post
	}{
		{
			name: "full post",
			payload: map[string]interface{}{
				"id":           "abc123",
				"name":         "t3_abc123",
				"title":        "Hello World",
				"subreddit":    "golang",
				"author":       "testuser",
				"score":        float64(42),
				"num_comments": float64(5),
				"created_utc":  float64(1700000000),
				"permalink":    "/r/golang/comments/abc123/hello_world/",
				"url":          "https://example.com",
				"selftext":     "body text",
				"is_self":      true,
				"over_18":      false,
				"is_video":     false,
				"stickied":     true,
			},
			want: Post{
				ID: "abc123", Name: "t3_abc123", Title: "Hello World",
				Subreddit: "golang", Author: "testuser", Score: 42,
				NumComments: 5, CreatedUTC: 1700000000,
				Permalink: "/r/golang/comments/abc123/hello_world/",
				URL: "https://example.com", Selftext: "body text",
				IsSelf: true, Over18: false, IsVideo: false, Stickied: true,
			},
		},
		{
			name:    "empty payload",
			payload: map[string]interface{}{},
			want: Post{
				IsSelf: true, // default
			},
		},
		{
			name: "score as int",
			payload: map[string]interface{}{
				"id":    "x",
				"score": float64(1234),
			},
			want: Post{ID: "x", Score: 1234, IsSelf: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePost(tt.payload)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePost() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseListing(t *testing.T) {
	after := "cursor123"
	tests := []struct {
		name string
		data map[string]interface{}
		want ListingPage
	}{
		{
			name: "normal listing",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"children": []interface{}{
						map[string]interface{}{
							"data": map[string]interface{}{
								"id":    "p1",
								"title": "Post 1",
							},
						},
						map[string]interface{}{
							"data": map[string]interface{}{
								"id":    "p2",
								"title": "Post 2",
							},
						},
					},
					"after": "cursor123",
				},
			},
			want: ListingPage{
				Items: []Post{
					{ID: "p1", Title: "Post 1", IsSelf: true},
					{ID: "p2", Title: "Post 2", IsSelf: true},
				},
				After: &after,
			},
		},
		{
			name: "empty listing",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"children": []interface{}{},
				},
			},
			want: ListingPage{Items: []Post{}},
		},
		{
			name: "nil data",
			data: map[string]interface{}{},
			want: ListingPage{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseListing(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseListing() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseUserProfile(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		want UserProfile
	}{
		{
			name: "with data wrapper",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"name":          "johndoe",
					"link_karma":    float64(1000),
					"comment_karma": float64(2000),
					"created_utc":   float64(1600000000),
					"is_gold":       true,
					"is_mod":        false,
				},
			},
			want: UserProfile{
				Name: "johndoe", LinkKarma: 1000, CommentKarma: 2000,
				CreatedUTC: 1600000000, IsGold: true, IsMod: false,
			},
		},
		{
			name: "flat data (no wrapper)",
			data: map[string]interface{}{
				"name":          "jane",
				"link_karma":    float64(500),
				"comment_karma": float64(100),
			},
			want: UserProfile{
				Name: "jane", LinkKarma: 500, CommentKarma: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseUserProfile(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUserProfile() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
