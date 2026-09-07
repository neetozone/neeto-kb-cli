package commands

import (
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Manage categories",
}

var categoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/categories", paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "categories", []output.Breadcrumb{
			{Label: "List articles in category", Command: binaryName() + " articles list --category-id <id>"},
		})
		return nil
	},
}

func init() {
	addPaginationFlags(categoriesListCmd)

	categoriesCmd.AddCommand(categoriesListCmd)
	register(func(root *cobra.Command) { root.AddCommand(categoriesCmd) })
}
