package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sammyne/reddit-cli/pkg/reddit"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Reddit posts",
	Long: `Search Reddit posts

Examples:
  reddit search "python async"
  reddit search "rust vs go" -r programming --sort top --time year`,
	Args: cobra.ExactArgs(1),
	Run:  runSearch,
}

func init() {
	searchCmd.Flags().StringP("subreddit", "r", "", "Search within subreddit")
	searchCmd.Flags().StringP("sort", "s", "relevance", "Sort order (relevance/hot/top/new/comments)")
	searchCmd.Flags().StringP("time", "t", "all", "Time filter (hour/day/week/month/year/all)")
	searchCmd.Flags().IntP("limit", "n", 25, "Number of results")
	searchCmd.Flags().String("after", "", "Pagination cursor")
	searchCmd.Flags().Bool("json", false, "Output as JSON")
	searchCmd.Flags().Bool("yaml", false, "Output as YAML")
	searchCmd.Flags().BoolP("compact", "c", false, "Compact output (fewer fields)")
	searchCmd.Flags().Bool("full-text", false, "Show full title without truncation")
	searchCmd.Flags().StringP("output", "o", "", "Save output to file")

	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) {
	query := args[0]
	subreddit, _ := cmd.Flags().GetString("subreddit")
	sortBy, _ := cmd.Flags().GetString("sort")
	timeFilter, _ := cmd.Flags().GetString("time")
	limit, _ := cmd.Flags().GetInt("limit")
	after, _ := cmd.Flags().GetString("after")
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")
	compact, _ := cmd.Flags().GetBool("compact")
	fullText, _ := cmd.Flags().GetBool("full-text")
	outputFile, _ := cmd.Flags().GetString("output")

	cred := reddit.GetCredential() // optional auth

	client := reddit.NewRedditClient(cred)
	client.Open()
	defer client.Close()

	data, err := client.Search(query, subreddit, sortBy, timeFilter, limit, after)
	if err != nil {
		exitForError(err, asJSON, asYAML, "Search failed")
		return
	}

	listing := reddit.ParseListing(data)
	posts := listing.Items

	if len(posts) > 0 {
		reddit.SaveIndex(posts, "search:"+query)
	}

	// --output: save to file
	if outputFile != "" {
		outData := interface{}(data)
		if compact {
			outData = postsToMaps(posts)
		}
		if err := reddit.SaveOutputToFile(outData, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "\033[31m❌ Failed to save: %v\033[0m\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "\033[32m✅ Saved to %s\033[0m\n", outputFile)
		return
	}

	// --compact
	outData := interface{}(data)
	localJSON, localYAML := asJSON, asYAML
	if compact {
		outData = postsToMaps(posts)
		if !localJSON && !localYAML {
			localYAML = true
		}
	}

	// --json/--yaml
	if reddit.MaybePrintStructured(outData, localJSON, localYAML) {
		if listing.After != nil {
			fmt.Fprintf(os.Stderr, "  \033[2m▸ More: reddit search \"%s\" --after %s\033[0m\n", query, *listing.After)
		}
		return
	}

	// Rich terminal table
	renderSearchTable(posts, query, fullText)
	if listing.After != nil {
		fmt.Fprintf(os.Stderr, "  \033[2m▸ More: reddit search \"%s\" --after %s\033[0m\n", query, *listing.After)
	}
}

func renderSearchTable(posts []reddit.Post, query string, fullText bool) {
	if len(posts) == 0 {
		fmt.Fprintf(os.Stderr, "\033[33mNo results for '%s'\033[0m\n", query)
		return
	}

	maxTitle := 45
	if fullText {
		maxTitle = 200
	}

	// Header
	fmt.Fprintf(os.Stderr, "\n  \033[1m🔍 Search: \"%s\" — %d results\033[0m\n\n", query, len(posts))
	fmt.Fprintf(os.Stderr, "  \033[2m%-3s %6s  %-15s  %-*s  %-12s  %5s\033[0m\n",
		"#", "Score", "Subreddit", maxTitle, "Title", "Author", "💬")
	fmt.Fprintf(os.Stderr, "  %s\n", strings.Repeat("─", 3+6+2+15+2+maxTitle+2+12+2+5))

	for i, post := range posts {
		title := post.Title
		if !fullText && len(title) > maxTitle {
			title = title[:maxTitle-3] + "..."
		}
		sub := "r/" + post.Subreddit
		if len(sub) > 15 {
			sub = sub[:15]
		}
		author := post.Author
		if len(author) > 12 {
			author = author[:12]
		}

		fmt.Fprintf(os.Stderr, "  \033[2m%-3d\033[0m \033[33m%6s\033[0m  \033[35m%-15s\033[0m  \033[1;36m%-*s\033[0m  \033[32m%-12s\033[0m  \033[2m%5d\033[0m\n",
			i+1,
			reddit.FormatScore(post.Score),
			sub,
			maxTitle, title,
			author,
			post.NumComments,
		)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  \033[2m💡 Use \033[1mreddit show <#>\033[0m\033[2m to read a result\033[0m")
}

func exitForError(err error, asJSON, asYAML bool, prefix string) {
	message := err.Error()
	if prefix != "" {
		message = prefix + ": " + message
	}
	code := reddit.ErrorCodeFor(err)
	if reddit.EmitError(code, message, asJSON, asYAML) {
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "\033[31m❌ [%s] %s\033[0m\n", code, message)
	os.Exit(1)
}

func postsToMaps(posts []reddit.Post) []map[string]interface{} {
	result := make([]map[string]interface{}, len(posts))
	for i, p := range posts {
		result[i] = map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"title":        p.Title,
			"subreddit":    p.Subreddit,
			"author":       p.Author,
			"score":        p.Score,
			"num_comments": p.NumComments,
			"permalink":    p.Permalink,
			"url":          p.URL,
			"created_utc":  p.CreatedUTC,
		}
	}
	return result
}
