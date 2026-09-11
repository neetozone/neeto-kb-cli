package commands

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var (
	htmlTagPattern = regexp.MustCompile(`<[^>]*>`)
	blankPattern   = regexp.MustCompile(`[\s\x00-\x08\x0b\x0e-\x1f\x7f-\x9f]+`)
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search articles",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if v, _ := cmd.Flags().GetString("search-term"); v != "" {
			params.Set("search_term", v)
		}

		data, err := c.Get("/search", params)
		if err != nil {
			return err
		}

		if rendersTable() {
			data = compactMatches(data)
		}

		printList(data, "matches", []output.Breadcrumb{
			{Label: "Show article", Command: "neetokb articles show <id>"},
		})
		return nil
	},
}

func compactMatches(data json.RawMessage) json.RawMessage {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(data, &payload); err != nil {
		return data
	}

	var matches []map[string]interface{}
	if err := json.Unmarshal(payload["matches"], &matches); err != nil {
		return data
	}

	compacted := make([]map[string]interface{}, len(matches))
	for i, match := range matches {
		compacted[i] = map[string]interface{}{
			"id":              match["id"],
			"matched_content": matchedSnippet(match),
			"url":             match["url"],
		}
	}

	encoded, err := json.Marshal(compacted)
	if err != nil {
		return data
	}
	payload["matches"] = encoded

	out, err := json.Marshal(payload)
	if err != nil {
		return data
	}
	return out
}

func matchedSnippet(match map[string]interface{}) interface{} {
	for _, field := range []string{"matched_content", "matched_title", "matched_category"} {
		if snippet := plainText(match[field]); snippet != "" {
			return snippet
		}
	}
	return nil
}

func plainText(value interface{}) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}

	text = html.UnescapeString(htmlTagPattern.ReplaceAllString(text, ""))
	return strings.TrimSpace(blankPattern.ReplaceAllString(text, " "))
}

func init() {
	addPaginationFlags(searchCmd)
	searchCmd.Flags().String("search-term", "", "Search query")
	_ = searchCmd.MarkFlagRequired("search-term")

	register(func(root *cobra.Command) { root.AddCommand(searchCmd) })
}
