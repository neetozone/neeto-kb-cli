package commands

import (
	"fmt"
	"net/http"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

func articleAction(use, short, method, pathSuffix string, breadcrumbs []output.Breadcrumb) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := send(cmd, method, fmt.Sprintf("/articles/%s%s", args[0], pathSuffix), nil)
			if err != nil {
				return err
			}

			printActionResult(data, breadcrumbs)
			return nil
		},
	}
}

var articlesTrashCmd = articleAction(
	"trash", "Move an article to the trash, where it stays for 30 days before it is deleted for good; restore it with trash restore",
	http.MethodDelete, "", []output.Breadcrumb{{Label: "Trash", Command: "neetokb trash list"}},
)

var articlesArchiveCmd = articleAction(
	"archive", "Archive an article: it becomes a draft that is kept out of the public knowledge base",
	http.MethodPost, "/archive", []output.Breadcrumb{{Label: "Unarchive", Command: "neetokb articles unarchive <id>"}},
)

var articlesUnarchiveCmd = articleAction(
	"unarchive", "Bring an archived article back as a draft", http.MethodDelete, "/archive", nil,
)

var articlesDiscardDraftCmd = articleAction(
	"discard-draft", "Throw away the unpublished draft changes of an article and keep its live version", http.MethodDelete, "/draft", nil,
)

var articlesCloneCmd = articleAction(
	"clone", "Copy an article into a new draft", http.MethodPost, "/clone",
	[]output.Breadcrumb{{Label: "Show", Command: "neetokb articles show <id>"}},
)

var articlesBulkTrashCmd = &cobra.Command{
	Use:   "bulk-trash",
	Short: "Move up to 100 articles to the trash; if one cannot be trashed none are",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodDelete, "/articles/bulk_destroy", flagValues(cmd, "ids[]=article_ids"))
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{{Label: "Trash", Command: "neetokb trash list"}})
		return nil
	},
}

var articlesBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Move up to 100 articles to a category, set their state, or archive or unarchive them; give exactly one of --category-id, --state or --status",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := flagValues(cmd, "ids[]=article_ids", "category-id", "state", "status")
		data, err := send(cmd, http.MethodPatch, "/articles/bulk_update", body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var trashCmd = &cobra.Command{
	Use:   "trash",
	Short: "Manage trashed articles",
}

var trashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List trashed articles and the days left before each is deleted for good",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/trashed_articles", paginationParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "trashed_articles", []output.Breadcrumb{
			{Label: "Restore", Command: "neetokb trash restore <id>"},
			{Label: "Delete forever", Command: "neetokb trash delete <id>"},
		})
		return nil
	},
}

var trashRestoreCmd = &cobra.Command{
	Use:   "restore <id>",
	Short: "Restore a trashed article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, fmt.Sprintf("/trashed_articles/%s/restoration", args[0]), nil)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{{Label: "Show", Command: "neetokb articles show <id>"}})
		return nil
	},
}

var trashDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a trashed article for good; this cannot be undone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/trashed_articles/%s", args[0]), nil, "Article deleted for good.")
	},
}

func init() {
	articlesBulkTrashCmd.Flags().String("ids", "", "Article IDs, slugs or permalink identifiers (comma-separated, up to 100)")
	markFlagsRequired(articlesBulkTrashCmd, "ids")

	articlesBulkUpdateCmd.Flags().String("ids", "", "Article IDs, slugs or permalink identifiers (comma-separated, up to 100)")
	articlesBulkUpdateCmd.Flags().String("category-id", "", "Move the articles to this category ID")
	articlesBulkUpdateCmd.Flags().String("state", "", "Set the state: draft or published")
	articlesBulkUpdateCmd.Flags().String("status", "", "Set the status: archive or active")
	markFlagsRequired(articlesBulkUpdateCmd, "ids")
	articlesBulkUpdateCmd.MarkFlagsOneRequired("category-id", "state", "status")
	articlesBulkUpdateCmd.MarkFlagsMutuallyExclusive("category-id", "state", "status")

	addPaginationFlags(trashListCmd)
	trashCmd.AddCommand(trashListCmd, trashRestoreCmd, trashDeleteCmd)

	articlesCmd.AddCommand(articlesTrashCmd, articlesBulkTrashCmd, articlesArchiveCmd, articlesUnarchiveCmd, articlesDiscardDraftCmd, articlesCloneCmd, articlesBulkUpdateCmd)
	register(func(root *cobra.Command) { root.AddCommand(trashCmd) })
}
