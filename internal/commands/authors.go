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

		data, err := c.Get("/authors", paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "authors", []output.Breadcrumb{
			{Label: "List articles", Command: binaryName() + " articles list"},
		})
		return nil
	},
}

func init() {
	addPaginationFlags(authorsListCmd)

	authorsCmd.AddCommand(authorsListCmd)
	register(func(root *cobra.Command) { root.AddCommand(authorsCmd) })
}
