package commands

import (
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var authorsCmd = &cobra.Command{
	Use:   "authors",
	Short: "List article authors",
}

var authorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List authors",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		sortBy, _ := cmd.Flags().GetString("sort-by")
		orderBy, _ := cmd.Flags().GetString("order-by")
		if err := validateSortOptions(sortBy, orderBy, []string{"name", "articles_count"}); err != nil {
			return err
		}
		if sortBy != "" {
			params.Set("sort_by", sortBy)
		}
		if orderBy != "" {
			params.Set("order_by", orderBy)
		}

		data, err := c.Get("/authors", params)
		if err != nil {
			return err
		}

		printList(data, "authors", []output.Breadcrumb{
			{Label: "List articles", Command: "neetokb articles list"},
		})
		return nil
	},
}

func init() {
	addPaginationFlags(authorsListCmd)
	authorsListCmd.Flags().String("sort-by", "", "Sort field: name or articles_count")
	authorsListCmd.Flags().String("order-by", "", "Sort order: asc or desc")

	authorsCmd.AddCommand(authorsListCmd)
	register(func(root *cobra.Command) { root.AddCommand(authorsCmd) })
}
