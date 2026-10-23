package commands

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var tagFlags = []string{"name", "description", "style"}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage article tags",
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags with the number of articles that use them",
	RunE: func(cmd *cobra.Command, args []string) error {
		params := queryValues(cmd, paginationParams(cmd), "search-term=search_term", "sort-by=sort_by", "order-by=order_by")
		data, err := fetch(cmd, "/tags", params)
		if err != nil {
			return err
		}

		printList(data, "tags", []output.Breadcrumb{
			{Label: "Update", Command: "neetokb tags update <id>"},
			{Label: "Merge", Command: "neetokb tags merge --primary-id <id> --secondary-ids <ids>"},
		})
		return nil
	},
}

var tagsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a tag",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, "/tags", flagValues(cmd, tagFlags...))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var tagsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/tags/%s", args[0]), flagValues(cmd, tagFlags...))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var tagsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a tag; it is removed from every article and the articles are kept",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/tags/%s", args[0]), url.Values{}, "Tag deleted.")
	},
}

var tagsMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge tags into one; articles of the secondary tags get the primary tag, which survives, and the secondary tags are deleted",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, "/tags/merge", flagValues(cmd, "primary-id=primary_id", "secondary-ids[]=secondary_ids"))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(tagsListCmd)
	tagsListCmd.Flags().String("search-term", "", "Show tags whose name contains this text")
	tagsListCmd.Flags().String("sort-by", "", "Sort field: name or articles_count")
	tagsListCmd.Flags().String("order-by", "", "Sort order: asc or desc")

	for _, cmd := range []*cobra.Command{tagsCreateCmd, tagsUpdateCmd} {
		cmd.Flags().String("name", "", "Tag name, unique ignoring case")
		cmd.Flags().String("description", "", "Tag description")
		cmd.Flags().String("style", "", "Tag colour: primary, secondary, info, success, warning or danger")
	}
	markFlagsRequired(tagsCreateCmd, "name")

	tagsMergeCmd.Flags().String("primary-id", "", "ID of the tag that survives")
	tagsMergeCmd.Flags().String("secondary-ids", "", "IDs of the tags merged into the primary tag and deleted (comma-separated)")
	markFlagsRequired(tagsMergeCmd, "primary-id", "secondary-ids")

	tagsCmd.AddCommand(tagsListCmd, tagsCreateCmd, tagsUpdateCmd, tagsDeleteCmd, tagsMergeCmd)
	register(func(root *cobra.Command) { root.AddCommand(tagsCmd) })
}
