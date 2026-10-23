package commands

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var recommendationFlags = []string{"title", "url", "matching-criteria=matching_criteria"}

func parseEntities(values []string) ([]map[string]string, error) {
	entities := make([]map[string]string, 0, len(values))
	for _, value := range values {
		kind, id, found := strings.Cut(value, ":")
		if !found || id == "" || (kind != "Article" && kind != "Category") {
			return nil, fmt.Errorf("invalid entity %q: use Article:<id> or Category:<id>", value)
		}
		entities = append(entities, map[string]string{"type": kind, "id": id})
	}
	return entities, nil
}

func recommendationBody(cmd *cobra.Command) (map[string]interface{}, error) {
	body := flagValues(cmd, recommendationFlags...)

	values, _ := cmd.Flags().GetStringArray("entity")
	clearAll, _ := cmd.Flags().GetBool("clear-entities")
	if len(values) > 0 && clearAll {
		return nil, fmt.Errorf("give either --entity or --clear-entities")
	}
	if len(values) > 0 || clearAll {
		entities, err := parseEntities(values)
		if err != nil {
			return nil, err
		}
		body["entities"] = entities
	}
	return body, nil
}

var recommendationsCmd = &cobra.Command{
	Use:   "recommendations",
	Short: "Manage recommendations",
}

var recommendationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recommendations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if v, _ := cmd.Flags().GetString("match-uri"); v != "" {
			params.Set("match_uri", v)
		}

		data, err := c.Get("/recommendations", params)
		if err != nil {
			return err
		}

		printList(data, "recommendations", []output.Breadcrumb{
			{Label: "Show", Command: "neetokb recommendations show <id>"},
			{Label: "List articles", Command: "neetokb articles list"},
		})
		return nil
	},
}

var recommendationsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a recommendation with the articles and categories it points to",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/recommendations/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetokb recommendations update <id>"},
		})
		return nil
	},
}

var recommendationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a recommendation that suggests articles or categories on pages matching a URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := recommendationBody(cmd)
		if err != nil {
			return err
		}

		data, err := send(cmd, http.MethodPost, "/recommendations", body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var recommendationsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a recommendation; --entity replaces all its entities and --clear-entities removes them",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := recommendationBody(cmd)
		if err != nil {
			return err
		}

		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/recommendations/%s", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var recommendationsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a recommendation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/recommendations/%s", args[0]), url.Values{}, "Recommendation deleted.")
	},
}

func init() {
	addPaginationFlags(recommendationsListCmd)
	recommendationsListCmd.Flags().String("match-uri", "", "Filter recommendations by URI")

	for _, cmd := range []*cobra.Command{recommendationsCreateCmd, recommendationsUpdateCmd} {
		cmd.Flags().String("title", "", "Recommendation title, unique ignoring case")
		cmd.Flags().String("url", "", "URL or regular expression the recommendation applies to")
		cmd.Flags().String("matching-criteria", "", "How the URL is matched: contains, exact_match or regex")
		cmd.Flags().StringArray("entity", nil, "Article or category to recommend as Article:<id> or Category:<id>; repeat for several")
	}
	recommendationsUpdateCmd.Flags().Bool("clear-entities", false, "Remove all entities from the recommendation")
	requireFlags(recommendationsCreateCmd, "title", "url")

	recommendationsCmd.AddCommand(recommendationsListCmd, recommendationsShowCmd, recommendationsCreateCmd, recommendationsUpdateCmd, recommendationsDeleteCmd)
	register(func(root *cobra.Command) { root.AddCommand(recommendationsCmd) })
}
