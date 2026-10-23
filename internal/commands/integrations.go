package commands

import (
	"net/http"

	"github.com/spf13/cobra"
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Manage integrations",
}

var integrationsNeetoChatCmd = &cobra.Command{
	Use:   "neeto-chat",
	Short: "Turn the NeetoChat integration on or off",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPatch, "/integrations/neeto_chat", flagValues(cmd, "enabled=neeto_chat_enabled"))
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	integrationsNeetoChatCmd.Flags().Bool("enabled", false, "Turn NeetoChat on (--enabled=false turns it off)")
	markFlagsRequired(integrationsNeetoChatCmd, "enabled")

	integrationsCmd.AddCommand(integrationsNeetoChatCmd)
	register(func(root *cobra.Command) { root.AddCommand(integrationsCmd) })
}
