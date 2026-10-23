package commands

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var articlesSchedulesCmd = &cobra.Command{
	Use:   "schedules",
	Short: "Manage when an article is published or unpublished automatically",
}

var articlesSchedulesListCmd = &cobra.Command{
	Use:   "list <article-id>",
	Short: "List the schedules of an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/articles/%s/schedules", args[0]), nil)
		if err != nil {
			return err
		}

		printList(data, "schedules", []output.Breadcrumb{
			{Label: "Create", Command: "neetokb articles schedules create <article-id> --kind publish --datetime <time> --time-zone <zone>"},
			{Label: "Delete", Command: "neetokb articles schedules delete <article-id> <schedule-id>"},
		})
		return nil
	},
}

var articlesSchedulesCreateCmd = &cobra.Command{
	Use:   "create <article-id>",
	Short: "Schedule an article to be published or unpublished at a future time; when both kinds exist, the unpublish time must come before the publish time",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := flagValues(cmd, "kind", "datetime", "time-zone", "publish-note")
		data, err := send(cmd, http.MethodPost, fmt.Sprintf("/articles/%s/schedules", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var articlesSchedulesDeleteCmd = &cobra.Command{
	Use:   "delete <article-id> <schedule-id>",
	Short: "Delete a schedule",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/articles/%s/schedules/%s", args[0], args[1]), url.Values{}, "Schedule deleted.")
	},
}

var articlesSeoCmd = &cobra.Command{
	Use:   "seo",
	Short: "Manage the search engine settings of an article",
}

var articlesSeoShowCmd = &cobra.Command{
	Use:   "show <article-id>",
	Short: "Show the SEO settings of an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/articles/%s/seo_setting", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetokb articles seo update <article-id>"},
		})
		return nil
	},
}

var articlesSeoUpdateCmd = &cobra.Command{
	Use:   "update <article-id>",
	Short: "Update the SEO settings of an article; only the given fields change",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := flagValues(cmd, "title", "title-suffix", "description", "keywords[]", "allow-search-engine-crawl")
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/articles/%s/seo_setting", args[0]), body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var articlesRevisionsCmd = &cobra.Command{
	Use:   "revisions",
	Short: "Browse and restore earlier versions of an article",
}

var articlesRevisionsListCmd = &cobra.Command{
	Use:   "list <article-id>",
	Short: "List the revisions of an article, newest first",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/articles/%s/revisions", args[0]), paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "revisions", []output.Breadcrumb{
			{Label: "Show", Command: "neetokb articles revisions show <article-id> <revision-id>"},
			{Label: "Restore", Command: "neetokb articles revisions restore <article-id> <revision-id>"},
		})
		return nil
	},
}

var articlesRevisionsShowCmd = &cobra.Command{
	Use:   "show <article-id> <revision-id>",
	Short: "Show a revision with its title and content",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/articles/%s/revisions/%s", args[0], args[1]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var articlesRevisionsRestoreCmd = &cobra.Command{
	Use:   "restore <article-id> <revision-id>",
	Short: "Copy a revision into the draft of an article; the live content does not change until the draft is published",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, fmt.Sprintf("/articles/%s/revisions/%s/restoration", args[0], args[1]), nil)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	articlesSchedulesCreateCmd.Flags().String("kind", "", "What happens at the time: publish or unpublish")
	articlesSchedulesCreateCmd.Flags().String("datetime", "", "Future date and time (ISO 8601); read in --time-zone when it has no offset")
	articlesSchedulesCreateCmd.Flags().String("time-zone", "", "IANA time zone name, for example Asia/Kolkata")
	articlesSchedulesCreateCmd.Flags().String("publish-note", "", "Note saved with the publication")
	requireFlags(articlesSchedulesCreateCmd, "kind", "datetime", "time-zone")

	articlesSeoUpdateCmd.Flags().String("title", "", "Page title")
	articlesSeoUpdateCmd.Flags().String("title-suffix", "", "Text added after the page title")
	articlesSeoUpdateCmd.Flags().String("description", "", "Meta description")
	articlesSeoUpdateCmd.Flags().String("keywords", "", "Keywords (comma-separated); replaces the current list")
	articlesSeoUpdateCmd.Flags().Bool("allow-search-engine-crawl", false, "Let search engines crawl the article (--allow-search-engine-crawl=false blocks them)")

	addPaginationFlags(articlesRevisionsListCmd)

	articlesSchedulesCmd.AddCommand(articlesSchedulesListCmd, articlesSchedulesCreateCmd, articlesSchedulesDeleteCmd)
	articlesSeoCmd.AddCommand(articlesSeoShowCmd, articlesSeoUpdateCmd)
	articlesRevisionsCmd.AddCommand(articlesRevisionsListCmd, articlesRevisionsShowCmd, articlesRevisionsRestoreCmd)
	articlesCmd.AddCommand(articlesSchedulesCmd, articlesSeoCmd, articlesRevisionsCmd)
}
