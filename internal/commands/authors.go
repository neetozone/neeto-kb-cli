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
		if v, _ := cmd.Flags().GetString("sort-by"); v != "" {
			params.Set("sort_by", v)
		}
		if v, _ := cmd.Flags().GetString("order-by"); v != "" {
			params.Set("order_by", v)
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
