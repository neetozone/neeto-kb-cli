package commands

import (
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspace",
}

var workspaceInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show workspace information",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/setting", nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceInfoCmd)
	rootCmd.AddCommand(workspaceCmd)
}
