package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

func readArticleBatch(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %s: %w", path, err)
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("Invalid JSON in %s: %w", path, err)
	}

	switch value := parsed.(type) {
	case []any:
		return map[string]any{"articles": value}, nil
	case map[string]any:
		if articles, ok := value["articles"].([]any); ok {
			return map[string]any{"articles": articles}, nil
		}
	}
	return nil, fmt.Errorf(`%s must hold a JSON array of articles or an object such as {"articles":[...]}`, path)
}

func runArticleBatch(method, endpoint string, breadcrumbs []output.Breadcrumb) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("json-file")
		body, err := readArticleBatch(path)
		if err != nil {
			return err
		}

		data, err := send(cmd, method, endpoint, body)
		if err != nil {
			return err
		}

		printActionResult(data, breadcrumbs)
		return nil
	}
}

var articlesBatchCreateCmd = &cobra.Command{
	Use:   "batch-create",
	Short: "Create 1 to 25 articles from a JSON file, all or nothing; items are saved in order, so a later item can relate to an earlier item's slug",
	Long: `Create 1 to 25 articles in one call from a JSON file holding an array of articles or {"articles":[...]}.

Each item needs title, html_content and category (an array of category names, innermost first). It may also
have state (draft or published, default draft), slug, tags (array), related_article_ids (slugs, permalink
identifiers or UUIDs) and note.

The batch is all or nothing: if one item fails, no article is saved and the error names the item, counting
from 1. Items are saved in order, so related_article_ids may name the slug of an earlier item.

This differs from "articles bulk-update", which applies one shared change to many existing articles.`,
	RunE: runArticleBatch(http.MethodPost, "/articles/batch_create", []output.Breadcrumb{
		{Label: "Show", Command: "neetokb articles show <id>"},
	}),
}

var articlesBatchUpdateCmd = &cobra.Command{
	Use:   "batch-update",
	Short: "Update 1 to 25 articles from a JSON file, each with its own changes, all or nothing; an article may appear only once",
	Long: `Update 1 to 25 articles in one call from a JSON file holding an array of items or {"articles":[...]}.

Each item needs article_id (slug, permalink identifier or UUID) and at least one of title, html_content, slug,
state, category (array of names, innermost first), tags, related_article_ids and note.

The batch is all or nothing: if one item fails, nothing is saved and the error names the item, counting from 1.
Items are saved in order, so related_article_ids may name a slug set by an earlier item. An article may appear
only once in a batch.

This differs from "articles bulk-update", which applies one shared change (category, state or status) to many articles.`,
	RunE: runArticleBatch(http.MethodPatch, "/articles/batch_update", []output.Breadcrumb{
		{Label: "Show", Command: "neetokb articles show <id>"},
	}),
}

func init() {
	articlesBatchCreateCmd.Flags().String("json-file", "", `JSON file holding an array of articles or {"articles":[{"title":"...","html_content":"<p>...</p>","category":["Guides","Setup"]}]}`)
	markFlagsRequired(articlesBatchCreateCmd, "json-file")

	articlesBatchUpdateCmd.Flags().String("json-file", "", `JSON file holding an array of changes or {"articles":[{"article_id":"my-slug","title":"New title"}]}`)
	markFlagsRequired(articlesBatchUpdateCmd, "json-file")

	articlesCmd.AddCommand(articlesBatchCreateCmd, articlesBatchUpdateCmd)
}
