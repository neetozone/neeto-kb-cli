package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var articlesUnlistedLinksCmd = &cobra.Command{
	Use:   "unlisted-links",
	Short: "Manage an article's unlisted (secret) share link",
}

var articlesUnlistedLinksGetCmd = &cobra.Command{
	Use:   "get <article-id>",
	Short: "Fetch an article's unlisted share link, creating a never-expiring one if none exists (article-id: slug, permalink identifier a-XXXXXXXX, or UUID; article must be published)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/articles/%s/unlisted_link", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Regenerate", Command: binaryName() + " articles unlisted-links regenerate <article-id>"},
		})
		return nil
	},
}

var articlesUnlistedLinksRegenerateCmd = &cobra.Command{
	Use:   "regenerate <article-id>",
	Short: "Regenerate an article's unlisted share link, invalidating the old URL immediately (article-id: slug, permalink identifier a-XXXXXXXX, or UUID; article must be published)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("expiration-type"); v != "" {
			body["expiration_type"] = v
		}
		if v, _ := cmd.Flags().GetString("expiration-date"); v != "" {
			body["expiration_date"] = v
		}

		data, err := c.Put(fmt.Sprintf("/articles/%s/unlisted_link", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Get", Command: binaryName() + " articles unlisted-links get <article-id>"},
		})
		return nil
	},
}

func init() {
	articlesUnlistedLinksRegenerateCmd.Flags().String(
		"expiration-type", "",
		"Expiration preset: never (default, no expiration), one_day, seven_days, thirty_days, or custom; preset dates are computed server-side and reject --expiration-date",
	)
	articlesUnlistedLinksRegenerateCmd.Flags().String(
		"expiration-date", "",
		"Exact expiry datetime (ISO8601, future). Required with --expiration-type custom; implies custom when --expiration-type is omitted; rejected alongside presets or never",
	)

	articlesUnlistedLinksCmd.AddCommand(articlesUnlistedLinksGetCmd)
	articlesUnlistedLinksCmd.AddCommand(articlesUnlistedLinksRegenerateCmd)
	articlesCmd.AddCommand(articlesUnlistedLinksCmd)
}
