package commands

import (
	"github.com/neetozone/neeto-kb-cli/internal/output"
	"github.com/spf13/cobra"
)

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
			{Label: "List articles", Command: "neetokb articles list"},
		})
		return nil
	},
}

func init() {
	addPaginationFlags(recommendationsListCmd)
	recommendationsListCmd.Flags().String("match-uri", "", "Filter recommendations by URI")

	recommendationsCmd.AddCommand(recommendationsListCmd)
	rootCmd.AddCommand(recommendationsCmd)
}
