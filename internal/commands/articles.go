package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var (
	articleStates     = []string{"draft", "published"}
	articleListStates = []string{"draft", "published", "archived", "all"}
)

func validateArticleState(state string, allowed []string) error {
	if state == "" || slices.Contains(allowed, state) {
		return nil
	}
	return fmt.Errorf("invalid state %q: use one of %s", state, strings.Join(allowed, ", "))
}

var articleSortFields = []string{"title", "slug", "state", "category_name", "views_count", "created_at", "updated_at"}

func validateSortOptions(sortBy, orderBy string, sortFields []string) error {
	if sortBy != "" && !slices.Contains(sortFields, sortBy) {
		return fmt.Errorf("invalid sort field %q: use one of %s", sortBy, strings.Join(sortFields, ", "))
	}
	if orderBy != "" && orderBy != "asc" && orderBy != "desc" {
		return fmt.Errorf("invalid order %q: use asc or desc", orderBy)
	}
	return nil
}

func articleBody(cmd *cobra.Command, title, slug, htmlContent, state, category string) map[string]interface{} {
	article := flagValues(cmd, "tags[]", "related-articles[]=related_article_ids", "note")
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

		sortBy, _ := cmd.Flags().GetString("sort-by")
		orderBy, _ := cmd.Flags().GetString("order-by")
		if err := validateSortOptions(sortBy, orderBy, articleSortFields); err != nil {
			return err
		}

		if v, _ := cmd.Flags().GetString("state"); v != "" {
			if err := validateArticleState(v, articleListStates); err != nil {
				return err
			}
			params.Set("state", v)
		}
		if v, _ := cmd.Flags().GetString("search-term"); v != "" {
			params.Set("search_term", v)
		}
		if sortBy != "" {
			params.Set("sort_by", sortBy)
		}
		if orderBy != "" {
			params.Set("order_by", orderBy)
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
		category, _ := cmd.Flags().GetString("category")

		if err := validateArticleState(state, articleStates); err != nil {
			return err
		}

		data, err := c.Post("/articles", articleBody(cmd, title, slug, htmlContent, state, category))
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

		if err := validateArticleState(state, articleStates); err != nil {
			return err
		}

		data, err := c.Patch(fmt.Sprintf("/articles/%s", args[0]), articleBody(cmd, title, slug, htmlContent, state, category))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(articlesListCmd)
	articlesListCmd.Flags().String("state", "", "Filter by state: draft, published, archived or all (default: published)")
	articlesListCmd.Flags().String("search-term", "", "Search term")
	articlesListCmd.Flags().String("sort-by", "", "Sort field: title, slug, state, category_name, views_count, created_at, updated_at")
	articlesListCmd.Flags().String("order-by", "", "Sort order: asc or desc")
	articlesListCmd.Flags().String("category-id", "", "Filter by category ID")

	articlesCreateCmd.Flags().String("title", "", "Article title")
	articlesCreateCmd.Flags().String("slug", "", "Article slug")
	articlesCreateCmd.Flags().String("html-content", "", "Article HTML content")
	articlesCreateCmd.Flags().String("state", "", "Article state (draft/published)")
	articlesCreateCmd.Flags().String("category", "", "Category path, most specific first (comma-separated, e.g. 'Installation,Getting Started')")
	articlesCreateCmd.Flags().String("tags", "", "Tag names (comma-separated); missing tags are created")
	articlesCreateCmd.Flags().String("related-articles", "", "Related article IDs, slugs or permalink identifiers (comma-separated)")
	articlesCreateCmd.Flags().String("note", "", "Internal note")
	_ = articlesCreateCmd.MarkFlagRequired("category")

	articlesUpdateCmd.Flags().String("title", "", "Article title")
	articlesUpdateCmd.Flags().String("slug", "", "Article slug")
	articlesUpdateCmd.Flags().String("html-content", "", "Article HTML content")
	articlesUpdateCmd.Flags().String("state", "", "Article state (draft/published)")
	articlesUpdateCmd.Flags().String("category", "", "New category path, most specific first, comma-separated; missing categories are created")
	articlesUpdateCmd.Flags().String("tags", "", "Tag names (comma-separated) that replace all tags; missing tags are created, an empty value removes all")
	articlesUpdateCmd.Flags().String("related-articles", "", "Related article IDs, slugs or permalink identifiers (comma-separated) that replace all related articles; an empty value removes all")
	articlesUpdateCmd.Flags().String("note", "", "Internal note; an empty value deletes it")

	articlesCmd.AddCommand(articlesListCmd)
	articlesCmd.AddCommand(articlesShowCmd)
	articlesCmd.AddCommand(articlesCreateCmd)
	articlesCmd.AddCommand(articlesUpdateCmd)
	register(func(root *cobra.Command) { root.AddCommand(articlesCmd) })
}
