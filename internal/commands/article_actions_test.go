package commands

import (
	"reflect"
	"testing"
)

func TestArticlesBulkUpdate_StatusIsSentWithTheArticleIDs(t *testing.T) {
	req, _, err := runCommand(t, `{"updated_articles":[]}`, "articles", "bulk-update", "--ids", "a1, a2", "--status", "archive")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "PATCH" || req.Path != "/api/external/v2/articles/bulk_update" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := map[string]interface{}{"article_ids": []interface{}{"a1", "a2"}, "status": "archive"}
	if !reflect.DeepEqual(req.Body, want) {
		t.Errorf("body = %v", req.Body)
	}
}

func TestArticlesBulkUpdate_CategoryWithStateIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "articles", "bulk-update", "--ids", "a1", "--category-id", "c1", "--state", "published")

	if err == nil {
		t.Error("expected an error")
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}

func TestArticlesBulkUpdate_NoChangeIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "articles", "bulk-update", "--ids", "a1")

	if err == nil {
		t.Error("expected an error")
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}
