package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var articlesCmd = &cobra.Command{
	Use:   "articles",
	Short: "Manage articles",
}

var articlesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List articles",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)

		if v, _ := cmd.Flags().GetString("state"); v != "" {
			params.Set("state", v)
		}
		if v, _ := cmd.Flags().GetString("search-term"); v != "" {
			params.Set("search_term", v)
		}
		if v, _ := cmd.Flags().GetString("sort-by"); v != "" {
			params.Set("sort_by", v)
		}
		if v, _ := cmd.Flags().GetString("order-by"); v != "" {
			params.Set("order_by", v)
		}
		if v, _ := cmd.Flags().GetString("category-id"); v != "" {
			params.Set("category_id", v)
		}

		data, err := c.Get("/articles", params)
		if err != nil {
			return err
		}

		printList(data, "articles", []output.Breadcrumb{
			{Label: "Show", Command: "neetokb articles show <id>"},
			{Label: "Update", Command: "neetokb articles update <id>"},
		})
		return nil
	},
}

var articlesShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/articles/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var articlesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an article",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		slug, _ := cmd.Flags().GetString("slug")
		htmlContent, _ := cmd.Flags().GetString("html-content")
		state, _ := cmd.Flags().GetString("state")
		categoryFlag, _ := cmd.Flags().GetString("category")
		category := cli.SplitCSV(categoryFlag)

		article := map[string]interface{}{
			"category": category,
		}
		if title != "" {
			article["title"] = title
		}
		if slug != "" {
			article["slug"] = slug
		}
		if htmlContent != "" {
			article["html_content"] = htmlContent
		}
		if state != "" {
			article["state"] = state
		}

		data, err := c.Post("/articles", map[string]interface{}{"article": article})
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetokb articles show <id>"},
		})
		return nil
	},
}

var articlesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		article := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			article["title"] = v
		}
		if v, _ := cmd.Flags().GetString("slug"); v != "" {
			article["slug"] = v
		}
		if v, _ := cmd.Flags().GetString("html-content"); v != "" {
			article["html_content"] = v
		}
		if v, _ := cmd.Flags().GetString("state"); v != "" {
			article["state"] = v
		}

		data, err := c.Patch(fmt.Sprintf("/articles/%s", args[0]), map[string]interface{}{"article": article})
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(articlesListCmd)
	articlesListCmd.Flags().String("state", "", "Filter by state (draft/published)")
	articlesListCmd.Flags().String("search-term", "", "Search term")
	articlesListCmd.Flags().String("sort-by", "", "Sort field")
	articlesListCmd.Flags().String("order-by", "", "Sort order (ASC/DESC)")
	articlesListCmd.Flags().String("category-id", "", "Filter by category ID")

	articlesCreateCmd.Flags().String("title", "", "Article title")
	articlesCreateCmd.Flags().String("slug", "", "Article slug")
	articlesCreateCmd.Flags().String("html-content", "", "Article HTML content")
	articlesCreateCmd.Flags().String("state", "", "Article state (draft/published)")
	articlesCreateCmd.Flags().String("category", "", "Category path, most specific first (comma-separated, e.g. 'Installation,Getting Started')")
	_ = articlesCreateCmd.MarkFlagRequired("category")

	articlesUpdateCmd.Flags().String("title", "", "Article title")
	articlesUpdateCmd.Flags().String("slug", "", "Article slug")
	articlesUpdateCmd.Flags().String("html-content", "", "Article HTML content")
	articlesUpdateCmd.Flags().String("state", "", "Article state (draft/published)")

	articlesCmd.AddCommand(articlesListCmd)
	articlesCmd.AddCommand(articlesShowCmd)
	articlesCmd.AddCommand(articlesCreateCmd)
	articlesCmd.AddCommand(articlesUpdateCmd)
	register(func(root *cobra.Command) { root.AddCommand(articlesCmd) })
}
