package cmd

import (
	"fmt"
	"os"

	"github.com/sammyne/reddit-cli/pkg/reddit"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Extract browser cookies for Reddit authentication",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if already logged in
		if cred := reddit.GetCredential(); cred != nil {
			fmt.Fprintln(os.Stderr, "\033[32m✅ Already authenticated\033[0m")
			return
		}
		fmt.Fprintln(os.Stderr, "\033[2m🔍 Searching for Reddit cookies in browsers...\033[0m")
		cred := reddit.ExtractBrowserCredential()
		if cred != nil {
			fmt.Fprintf(os.Stderr, "\033[32m✅ Login successful!\033[0m (%d cookies extracted)\n", len(cred.Cookies))
		} else {
			fmt.Fprintln(os.Stderr, "\033[31m❌ No Reddit cookies found.\033[0m")
			fmt.Fprintln(os.Stderr, "  \033[2mPlease login to reddit.com in your browser first, then retry.\033[0m")
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear saved Reddit cookies",
	Run: func(cmd *cobra.Command, args []string) {
		reddit.ClearCredential()
		fmt.Fprintln(os.Stderr, "\033[32m✅ Credentials cleared\033[0m")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Run:   runStatus,
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user profile (karma, account age)",
	Run:   runWhoami,
}

func init() {
	statusCmd.Flags().Bool("json", false, "Output as JSON")
	statusCmd.Flags().Bool("yaml", false, "Output as YAML")
	whoamiCmd.Flags().Bool("json", false, "Output as JSON")
	whoamiCmd.Flags().Bool("yaml", false, "Output as YAML")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(whoamiCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")

	cred := reddit.GetCredential()
	session := reddit.SessionFromCredential(cred)

	info := map[string]interface{}{
		"authenticated":   session.IsAuthenticated(),
		"cookie_count":    len(session.Cookies),
		"credential_file": reddit.CredentialFile(),
		"source":          session.Source,
		"username":        session.Username,
		"capabilities":    session.SortedCapabilities(),
		"modhash_present": session.Modhash != "",
		"error":           nilIfEmpty(session.ValidationError),
	}

	// Live validation if we have credentials
	if cred != nil {
		client := reddit.NewRedditClient(cred)
		client.Open()
		result := client.ValidateSession()
		client.Close()

		info["authenticated"] = result["authenticated"]
		if u, ok := result["username"].(string); ok && u != "" {
			info["username"] = u
		}
		if caps, ok := result["capabilities"]; ok {
			info["capabilities"] = caps
		}
		if mp, ok := result["modhash_present"]; ok {
			info["modhash_present"] = mp
		}
		info["error"] = nilIfEmpty(fmt.Sprint(result["error"]))
	}

	if reddit.MaybePrintStructured(info, asJSON, asYAML) {
		return
	}

	// Rich terminal output
	if info["authenticated"] == true {
		fmt.Fprintf(os.Stderr, "\033[32m✅ Authenticated\033[0m (%v cookies)\n", info["cookie_count"])
		if u, ok := info["username"].(string); ok && u != "" {
			fmt.Fprintf(os.Stderr, "  \033[2muser: %s\033[0m\n", u)
		}
		fmt.Fprintf(os.Stderr, "  \033[2mcapabilities: %s\033[0m\n", fmtCaps(info["capabilities"]))
		fmt.Fprintf(os.Stderr, "  \033[2msource: %s\033[0m\n", info["source"])
	} else {
		fmt.Fprintln(os.Stderr, "\033[33m⚠️  Not authenticated\033[0m")
		if e := info["error"]; e != nil && e != "" {
			fmt.Fprintf(os.Stderr, "  \033[2m%v\033[0m\n", e)
		}
		fmt.Fprintln(os.Stderr, "  \033[2mUse 'reddit login' to extract cookies from your browser\033[0m")
	}
}

func runWhoami(cmd *cobra.Command, args []string) {
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")

	cred := reddit.GetCredential()
	if cred == nil {
		fmt.Fprintln(os.Stderr, "\033[33m⚠️  Not logged in\033[0m. Use \033[1mreddit login\033[0m to authenticate")
		os.Exit(1)
	}

	client := reddit.NewRedditClient(cred)
	client.Open()
	defer client.Close()

	me, err := client.GetMe()
	if err != nil {
		code := reddit.ErrorCodeFor(err)
		if reddit.EmitError(code, err.Error(), asJSON, asYAML) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "\033[31m❌ [%s] %s\033[0m\n", code, err)
		os.Exit(1)
	}

	// Try to get full profile
	name := reddit.AsString(me["name"])
	if name == "" {
		name = client.Session.Username
	}
	profileData := me
	if name != "" {
		if about, err := client.GetUserAbout(name); err == nil {
			profileData = about
			profileData["name"] = name
		}
	}

	if reddit.MaybePrintStructured(profileData, asJSON, asYAML) {
		return
	}

	// Rich terminal output
	profile := reddit.ParseUserProfile(profileData)
	karmaPost := profile.LinkKarma
	karmaComment := profile.CommentKarma
	totalKarma := karmaPost + karmaComment

	gold := ""
	if profile.IsGold {
		gold = "⭐ "
	}
	mod := ""
	if profile.IsMod {
		mod = "🛡️ "
	}

	fmt.Fprintln(os.Stderr, "╭─── 👤 Me ───╮")
	fmt.Fprintf(os.Stderr, "│ \033[1;36mu/%s\033[0m %s%s\n", profile.Name, gold, mod)
	fmt.Fprintf(os.Stderr, "│ 📊 Total karma: %s\n", fmtNum(totalKarma))
	fmt.Fprintf(os.Stderr, "│    Post: %s · Comment: %s\n", fmtNum(karmaPost), fmtNum(karmaComment))
	fmt.Fprintf(os.Stderr, "│ 📅 Joined: %s\n", reddit.FormatTime(profile.CreatedUTC))
	fmt.Fprintln(os.Stderr, "╰──────────────╯")
}

func nilIfEmpty(s string) interface{} {
	if s == "" || s == "<nil>" {
		return nil
	}
	return s
}

func fmtCaps(v interface{}) string {
	switch caps := v.(type) {
	case []string:
		if len(caps) == 0 {
			return "-"
		}
		result := ""
		for i, c := range caps {
			if i > 0 {
				result += ", "
			}
			result += c
		}
		return result
	default:
		return fmt.Sprintf("%v", v)
	}
}

func fmtNum(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
