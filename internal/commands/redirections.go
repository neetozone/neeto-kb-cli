package commands

import (
	"fmt"
	"net/http"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var redirectionFlags = []string{"from-path", "to-path", "type=redirection_type"}

var redirectionsCmd = &cobra.Command{
	Use:   "redirections",
	Short: "Manage redirections",
}

var redirectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List redirections, newest first",
	RunE: func(cmd *cobra.Command, args []string) error {
		params := queryValues(cmd, paginationParams(cmd), "search-term=search_term")
		data, err := fetch(cmd, "/redirections", params)
		if err != nil {
			return err
		}

		printList(data, "redirections", []output.Breadcrumb{
			{Label: "Show", Command: "neetokb redirections show <id>"},
			{Label: "Update", Command: "neetokb redirections update <id>"},
		})
		return nil
	},
}

var redirectionsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a redirection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/redirections/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var redirectionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a redirection from an old path to a new path or URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, "/redirections", flagValues(cmd, redirectionFlags...))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var redirectionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a redirection; only the given fields change",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/redirections/%s", args[0]), flagValues(cmd, redirectionFlags...))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var redirectionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a redirection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/redirections/%s", args[0]), nil, "Redirection deleted.")
	},
}

func init() {
	addPaginationFlags(redirectionsListCmd)
	redirectionsListCmd.Flags().String("search-term", "", "Show redirections whose from or to path contains this text")

	for _, cmd := range []*cobra.Command{redirectionsCreateCmd, redirectionsUpdateCmd} {
		cmd.Flags().String("from-path", "", "Old path, or a bare origin such as https://old.example.com for a domain redirection")
		cmd.Flags().String("to-path", "", "New path or full URL")
		cmd.Flags().String("type", "", "Redirection type: page (default) or domain")
	}
	markFlagsRequired(redirectionsCreateCmd, "from-path", "to-path")

	redirectionsCmd.AddCommand(redirectionsListCmd, redirectionsShowCmd, redirectionsCreateCmd, redirectionsUpdateCmd, redirectionsDeleteCmd)
	register(func(root *cobra.Command) { root.AddCommand(redirectionsCmd) })
}
