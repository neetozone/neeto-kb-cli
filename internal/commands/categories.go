package commands

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var categoryFlags = []string{"name", "description", "private=is_private", "parent=parent_id", "position"}

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Manage categories",
}

var categoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if flat, _ := cmd.Flags().GetBool("flat"); flat {
			params.Set("flat", "true")
		}

		data, err := c.Get("/categories", params)
		if err != nil {
			return err
		}

		printList(data, "categories", []output.Breadcrumb{
			{Label: "Show", Command: "neetokb categories show <id>"},
			{Label: "List articles in category", Command: "neetokb articles list --category-id <id>"},
		})
		return nil
	},
}

var categoriesShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a category (id: UUID or slug)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/categories/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetokb categories update <id>"},
			{Label: "Order articles", Command: "neetokb categories reorder-articles <id> --article-ids <ids>"},
		})
		return nil
	},
}

var categoriesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a category",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, "/categories", flagValues(cmd, categoryFlags...))
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetokb categories show <id>"},
		})
		return nil
	},
}

var categoriesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Rename, move, reorder or change the privacy of a category; --parent \"\" moves it to the top level",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/categories/%s", args[0]), flagValues(cmd, categoryFlags...))
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var categoriesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a category; its articles and subcategories move to --move-to-category-id, which a top-level category with content needs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := queryValues(cmd, url.Values{}, "move-to-category-id=move_to_category_id")
		return removeAndConfirm(cmd, fmt.Sprintf("/categories/%s", args[0]), params, "Category deleted.")
	},
}

var categoriesReorderArticlesCmd = &cobra.Command{
	Use:   "reorder-articles <id>",
	Short: "Set the order of articles in a category; articles not listed follow in their current order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/categories/%s/article_order", args[0]), flagValues(cmd, "article-ids[]"))
		if err != nil {
			return err
		}

		printList(data, "articles", nil)
		return nil
	},
}

func init() {
	addPaginationFlags(categoriesListCmd)
	categoriesListCmd.Flags().Bool("flat", false, "List every category, including empty and private ones, as flat rows with counts")

	for _, cmd := range []*cobra.Command{categoriesCreateCmd, categoriesUpdateCmd} {
		cmd.Flags().String("name", "", "Category name (up to 120 characters)")
		cmd.Flags().String("description", "", "Category description")
		cmd.Flags().Bool("private", false, "Hide the category from visitors who are not signed in (--private=false makes it public)")
		cmd.Flags().String("parent", "", "Parent category ID or slug")
		cmd.Flags().Int("position", 0, "Position among sibling categories, starting at 1")
	}
	requireFlags(categoriesCreateCmd, "name")

	categoriesDeleteCmd.Flags().String("move-to-category-id", "", "Category ID or slug that receives the articles and subcategories")

	categoriesReorderArticlesCmd.Flags().String("article-ids", "", "Article IDs, slugs or permalink identifiers in the wanted order (comma-separated)")
	requireFlags(categoriesReorderArticlesCmd, "article-ids")

	categoriesCmd.AddCommand(categoriesListCmd, categoriesShowCmd, categoriesCreateCmd, categoriesUpdateCmd, categoriesDeleteCmd, categoriesReorderArticlesCmd)
	register(func(root *cobra.Command) { root.AddCommand(categoriesCmd) })
}
