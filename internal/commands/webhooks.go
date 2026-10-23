package commands

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var webhookFlags = []string{"endpoint", "secret", "active=is_active", "event=events"}

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhooks",
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/webhooks", paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "webhooks", []output.Breadcrumb{
			{Label: "Show", Command: "neetokb webhooks show <id>"},
			{Label: "Events", Command: "neetokb webhooks events"},
		})
		return nil
	},
}

var webhooksShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a webhook; its secret is never shown",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/webhooks/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook that sends article events to an endpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, "/webhooks", map[string]interface{}{"webhook": flagValues(cmd, webhookFlags...)})
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var webhooksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a webhook; --event replaces the whole event list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]interface{}{"webhook": flagValues(cmd, webhookFlags...)}
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/webhooks/%s", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/webhooks/%s", args[0]), url.Values{}, "Webhook deleted.")
	},
}

var webhooksEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "List the events a webhook can subscribe to",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/webhooks/events", nil)
		if err != nil {
			return err
		}

		printList(data, "events", nil)
		return nil
	},
}

func init() {
	addPaginationFlags(webhooksListCmd)

	for _, cmd := range []*cobra.Command{webhooksCreateCmd, webhooksUpdateCmd} {
		cmd.Flags().String("endpoint", "", "URL that receives the events")
		cmd.Flags().String("secret", "", "Signing secret; it can be set but never read back")
		cmd.Flags().Bool("active", false, "Send events (--active=false pauses the webhook)")
		cmd.Flags().StringArray("event", nil, "Event to send: create, update, delete, publish or unpublish; repeat for several")
	}
	requireFlags(webhooksCreateCmd, "endpoint")

	webhooksCmd.AddCommand(webhooksListCmd, webhooksShowCmd, webhooksCreateCmd, webhooksUpdateCmd, webhooksDeleteCmd, webhooksEventsCmd)
	register(func(root *cobra.Command) { root.AddCommand(webhooksCmd) })
}
