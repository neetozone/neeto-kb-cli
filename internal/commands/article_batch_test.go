package commands

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeBatchFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "batch.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestArticlesBatchCreate_BareArrayFileIsPostedWrappedInArticles(t *testing.T) {
	path := writeBatchFile(t, `[{"title":"Setup","html_content":"<p>Hi</p>","category":["Guides"]}]`)

	req, _, err := runCommand(t, `{"articles":[{"id":"1","slug":"setup","permalink":"https://x/p/a-1"}]}`, "articles", "batch-create", "--json-file", path)
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "POST" || req.Path != "/api/external/v2/articles/batch_create" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := map[string]interface{}{
		"articles": []interface{}{
			map[string]interface{}{"title": "Setup", "html_content": "<p>Hi</p>", "category": []interface{}{"Guides"}},
		},
	}
	if !reflect.DeepEqual(req.Body, want) {
		t.Errorf("body = %v", req.Body)
	}
}

func TestArticlesBatchUpdate_ObjectFileIsPatchedAsTheBody(t *testing.T) {
	path := writeBatchFile(t, `{"articles":[{"article_id":"setup","title":"New title"}]}`)

	req, _, err := runCommand(t, `{"articles":[{"id":"1","slug":"setup","state":"draft","title":"New title"}]}`, "articles", "batch-update", "--json-file", path)
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "PATCH" || req.Path != "/api/external/v2/articles/batch_update" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := map[string]interface{}{
		"articles": []interface{}{
			map[string]interface{}{"article_id": "setup", "title": "New title"},
		},
	}
	if !reflect.DeepEqual(req.Body, want) {
		t.Errorf("body = %v", req.Body)
	}
}
