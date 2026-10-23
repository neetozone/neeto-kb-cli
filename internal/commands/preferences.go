package commands

import (
	"net/http"

	"github.com/spf13/cobra"
)

var preferencesCmd = &cobra.Command{
	Use:   "preferences",
	Short: "Manage your own notification preferences; needs a signed-in user, not an API key",
}

var preferencesShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show your preferences",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/my_preferences", nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var preferencesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update your preferences",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]interface{}{"my_preferences": flagValues(cmd, "monthly-reports-mail")}
		data, err := send(cmd, http.MethodPatch, "/my_preferences", body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	preferencesUpdateCmd.Flags().Bool("monthly-reports-mail", false, "Get the monthly report email (--monthly-reports-mail=false turns it off)")

	preferencesCmd.AddCommand(preferencesShowCmd, preferencesUpdateCmd)
	register(func(root *cobra.Command) { root.AddCommand(preferencesCmd) })
}
