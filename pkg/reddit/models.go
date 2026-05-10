package reddit

// Post represents a Reddit post.
type Post struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Subreddit   string  `json:"subreddit"`
	Author      string  `json:"author"`
	Score       int     `json:"score"`
	NumComments int     `json:"num_comments"`
	CreatedUTC  float64 `json:"created_utc"`
	Permalink   string  `json:"permalink"`
	URL         string  `json:"url"`
	Selftext    string  `json:"selftext"`
	IsSelf      bool    `json:"is_self"`
	Over18      bool    `json:"over_18"`
	IsVideo     bool    `json:"is_video"`
	Stickied    bool    `json:"stickied"`
}

// ListingPage holds a page of listing results with pagination cursors.
type ListingPage struct {
	Items  []Post  `json:"items"`
	After  *string `json:"after"`
	Before *string `json:"before"`
}

// UserProfile holds basic user profile information.
type UserProfile struct {
	Name         string  `json:"name"`
	LinkKarma    int     `json:"link_karma"`
	CommentKarma int     `json:"comment_karma"`
	CreatedUTC   float64 `json:"created_utc"`
	IsGold       bool    `json:"is_gold"`
	IsMod        bool    `json:"is_mod"`
}
