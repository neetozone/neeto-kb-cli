package commands

import (
	"net/http"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var navigationLinksCmd = &cobra.Command{
	Use:   "navigation-links",
	Short: "Manage header and footer navigation links",
}

var navigationLinksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the header and footer navigation links",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/navigation_links", nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Replace", Command: "neetokb navigation-links update --json-file <file>"},
		})
		return nil
	},
}

var navigationLinksUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Replace the header links, footer links or both from a JSON file; a set that is sent becomes exactly that list and links left out are deleted",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("json-file")
		body, err := readJSONFile(path)
		if err != nil {
			return err
		}

		data, err := send(cmd, http.MethodPut, "/navigation_links", body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	navigationLinksUpdateCmd.Flags().String("json-file", "", `JSON file such as {"header_links":[{"label":"Docs","link":"https://example.com","kind":"external_link"}],"footer_links":[]}; each link has label, link and kind (external_link or action), and an optional id`)
	requireFlags(navigationLinksUpdateCmd, "json-file")

	navigationLinksCmd.AddCommand(navigationLinksListCmd, navigationLinksUpdateCmd)
	register(func(root *cobra.Command) { root.AddCommand(navigationLinksCmd) })
}
