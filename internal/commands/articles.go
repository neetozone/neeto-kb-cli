package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var articleListStates = []string{"draft", "published", "all"}

func validateArticleState(state string) error {
	if state == "" || slices.Contains(articleListStates, state) {
		return nil
	}
	return fmt.Errorf("invalid state %q: use one of %s", state, strings.Join(articleListStates, ", "))
}

func articleUpdateBody(title, slug, htmlContent, state, category string) map[string]interface{} {
	article := map[string]interface{}{}
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
	if category != "" {
		article["category"] = cli.SplitCSV(category)
	}
	return map[string]interface{}{"article": article}
}

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
			if err := validateArticleState(v); err != nil {
				return err
			}
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

		title, _ := cmd.Flags().GetString("title")
		slug, _ := cmd.Flags().GetString("slug")
		htmlContent, _ := cmd.Flags().GetString("html-content")
		state, _ := cmd.Flags().GetString("state")
		category, _ := cmd.Flags().GetString("category")

		data, err := c.Patch(fmt.Sprintf("/articles/%s", args[0]), articleUpdateBody(title, slug, htmlContent, state, category))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(articlesListCmd)
	articlesListCmd.Flags().String("state", "", "Filter by state: draft, published or all (default: published)")
	articlesListCmd.Flags().String("search-term", "", "Search term")
	articlesListCmd.Flags().String("sort-by", "", "Sort field: title, slug, state, category_name, views_count, created_at, updated_at")
	articlesListCmd.Flags().String("order-by", "", "Sort order: asc or desc")
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
	articlesUpdateCmd.Flags().String("category", "", "New category path, most specific first, comma-separated; missing categories are created")

	articlesCmd.AddCommand(articlesListCmd)
	articlesCmd.AddCommand(articlesShowCmd)
	articlesCmd.AddCommand(articlesCreateCmd)
	articlesCmd.AddCommand(articlesUpdateCmd)
	register(func(root *cobra.Command) { root.AddCommand(articlesCmd) })
}
