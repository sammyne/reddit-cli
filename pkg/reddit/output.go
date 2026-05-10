package reddit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

const schemaVersion = "1"

// SuccessPayload wraps data in the agent-friendly envelope.
func SuccessPayload(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"ok":             true,
		"schema_version": schemaVersion,
		"data":           data,
	}
}

// ErrorPayload wraps an error in the agent-friendly envelope.
func ErrorPayload(code, message string) map[string]interface{} {
	return map[string]interface{}{
		"ok":             false,
		"schema_version": schemaVersion,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}

// PrintJSON writes JSON to stdout.
func PrintJSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(data)
}

// PrintYAML writes YAML to stdout.
func PrintYAML(data interface{}) {
	out, err := yaml.Marshal(data)
	if err != nil {
		PrintJSON(data) // fallback
		return
	}
	fmt.Fprint(os.Stdout, string(out))
}

// ResolveOutputFormat determines the output format.
// Returns "json", "yaml", or "" (for terminal rendering).
func ResolveOutputFormat(asJSON, asYAML bool) string {
	if asJSON {
		return "json"
	}
	if asYAML {
		return "yaml"
	}
	env := strings.TrimSpace(strings.ToLower(os.Getenv("OUTPUT")))
	switch env {
	case "yaml":
		return "yaml"
	case "json":
		return "json"
	case "rich":
		return ""
	}
	if !IsTerminal() {
		return "yaml"
	}
	return ""
}

// MaybePrintStructured prints structured output if a machine-readable format
// is active. Returns true if output was printed.
func MaybePrintStructured(data interface{}, asJSON, asYAML bool) bool {
	f := ResolveOutputFormat(asJSON, asYAML)
	if f == "" {
		return false
	}
	payload := SuccessPayload(data)
	if f == "json" {
		PrintJSON(payload)
	} else {
		PrintYAML(payload)
	}
	return true
}

// EmitError prints a structured error if machine-readable format is active.
// Returns true if the error was emitted.
func EmitError(code, message string, asJSON, asYAML bool) bool {
	f := ResolveOutputFormat(asJSON, asYAML)
	if f == "" {
		return false
	}
	payload := ErrorPayload(code, message)
	if f == "json" {
		PrintJSON(payload)
	} else {
		PrintYAML(payload)
	}
	return true
}

// SaveOutputToFile saves structured output to a file, auto-detecting
// JSON or YAML by file extension.
func SaveOutputToFile(data interface{}, outputFile string) error {
	payload := SuccessPayload(data)
	ext := strings.ToLower(filepath.Ext(outputFile))

	var content []byte
	var err error
	if ext == ".yml" || ext == ".yaml" {
		content, err = yaml.Marshal(payload)
	} else {
		content, err = json.MarshalIndent(payload, "", "  ")
	}
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, content, 0o644)
}

// FormatScore formats a score as human-readable (e.g. 1.2k).
func FormatScore(score int) string {
	if score >= 1000 {
		return fmt.Sprintf("%.1fk", float64(score)/1000)
	}
	return fmt.Sprintf("%d", score)
}

// FormatTime formats a Unix timestamp as a relative time string.
func FormatTime(ts float64) string {
	if ts == 0 {
		return "-"
	}
	now := float64(time.Now().Unix())
	diff := now - ts
	if diff < 0 {
		return "just now"
	}
	if diff < 60 {
		return fmt.Sprintf("%ds ago", int(diff))
	}
	if diff < 3600 {
		return fmt.Sprintf("%dm ago", int(diff/60))
	}
	if diff < 86400 {
		return fmt.Sprintf("%dh ago", int(diff/3600))
	}
	if diff < 604800 {
		return fmt.Sprintf("%dd ago", int(diff/86400))
	}
	t := time.Unix(int64(ts), 0).UTC()
	return t.Format("2006-01-02")
}

// IsTerminal returns true if stdout is a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
