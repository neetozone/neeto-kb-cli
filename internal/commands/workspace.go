package commands

import (
	"net/http"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var workspaceSettingFlags = []string{
	"name=knowledge_base_name", "site-content-preference", "allow-signup=allow_user_signup",
	"show-article-nav-links", "header-text", "search-input-text", "copyright=copyright_content",
	"social-sharing=enable_social_media_sharing", "facebook-app-id",
}

var workspaceContactFlags = []string{"contact-title=title", "contact-kind=kind", "contact-link=link", "contact-button-label=button_label"}

var workspaceSeoFlags = []string{
	"title", "title-suffix", "description", "keywords[]", "allow-crawl=allow_search_engine_crawl",
	"google-analytics-id=google_analytics_tracking_id",
}

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspace",
}

var workspaceInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show workspace information",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/setting", nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetokb workspace update"},
			{Label: "SEO settings", Command: "neetokb workspace seo show"},
		})
		return nil
	},
}

var workspaceUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the knowledge base name, texts, sign-up, contact section and social sharing settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		setting := flagValues(cmd, workspaceSettingFlags...)
		if contact := flagValues(cmd, workspaceContactFlags...); len(contact) > 0 {
			setting["contact_section"] = contact
		}

		data, err := send(cmd, http.MethodPatch, "/setting", map[string]interface{}{"setting": setting})
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var workspaceSeoCmd = &cobra.Command{
	Use:   "seo",
	Short: "Manage the search engine settings of the whole knowledge base",
}

var workspaceSeoShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the knowledge base SEO settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/seo_setting", nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetokb workspace seo update"},
		})
		return nil
	},
}

var workspaceSeoUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the knowledge base SEO settings; --keywords replaces the whole list",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]interface{}{"seo_setting": flagValues(cmd, workspaceSeoFlags...)}
		data, err := send(cmd, http.MethodPatch, "/seo_setting", body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	workspaceUpdateCmd.Flags().String("name", "", "Knowledge base name (up to 80 characters)")
	workspaceUpdateCmd.Flags().String("site-content-preference", "", "What the home page shows, such as Categories")
	workspaceUpdateCmd.Flags().Bool("allow-signup", false, "Let visitors sign up (--allow-signup=false turns it off)")
	workspaceUpdateCmd.Flags().Bool("show-article-nav-links", false, "Show previous and next links on articles")
	workspaceUpdateCmd.Flags().String("header-text", "", "Home page heading")
	workspaceUpdateCmd.Flags().String("search-input-text", "", "Search box placeholder text")
	workspaceUpdateCmd.Flags().String("copyright", "", "Copyright text in the footer")
	workspaceUpdateCmd.Flags().Bool("social-sharing", false, "Enable social media sharing")
	workspaceUpdateCmd.Flags().String("facebook-app-id", "", "Facebook app ID (letters and digits)")
	workspaceUpdateCmd.Flags().String("contact-title", "", "Contact section title")
	workspaceUpdateCmd.Flags().String("contact-kind", "", "Contact section button kind, such as action")
	workspaceUpdateCmd.Flags().String("contact-link", "", "Contact section button link")
	workspaceUpdateCmd.Flags().String("contact-button-label", "", "Contact section button label")

	workspaceSeoUpdateCmd.Flags().String("title", "", "Page title")
	workspaceSeoUpdateCmd.Flags().String("title-suffix", "", "Text added after every page title")
	workspaceSeoUpdateCmd.Flags().String("description", "", "Meta description")
	workspaceSeoUpdateCmd.Flags().String("keywords", "", "Keywords (comma-separated); replaces the current list")
	workspaceSeoUpdateCmd.Flags().Bool("allow-crawl", false, "Let search engines crawl the knowledge base (--allow-crawl=false blocks them)")
	workspaceSeoUpdateCmd.Flags().String("google-analytics-id", "", "Google Analytics tracking ID")

	workspaceSeoCmd.AddCommand(workspaceSeoShowCmd, workspaceSeoUpdateCmd)
	workspaceCmd.AddCommand(workspaceInfoCmd, workspaceUpdateCmd, workspaceSeoCmd)
	register(func(root *cobra.Command) { root.AddCommand(workspaceCmd) })
}
