package commands

import (
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage workspace settings",
}

var settingsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show workspace settings",
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
	settingsCmd.AddCommand(settingsShowCmd)
	rootCmd.AddCommand(settingsCmd)
}
