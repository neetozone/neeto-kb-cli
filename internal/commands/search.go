package commands

import (
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
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

		printList(data, "matches", []output.Breadcrumb{
			{Label: "Show article", Command: "neetokb articles show <id>"},
		})
		return nil
	},
}

func init() {
	addPaginationFlags(searchCmd)
	searchCmd.Flags().String("search-term", "", "Search query")
	_ = searchCmd.MarkFlagRequired("search-term")

	register(func(root *cobra.Command) { root.AddCommand(searchCmd) })
}
