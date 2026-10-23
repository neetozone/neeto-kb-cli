package commands

import (
	"reflect"
	"testing"
)

func TestArticlesCreate_SendsArticleWithCategoryPath(t *testing.T) {
	req, out, err := runCommand(t, `{"article":{"id":"a1"}}`,
		"articles", "create", "--title", "Setup", "--slug", "setup", "--html-content", "<p>Hi</p>",
		"--state", "draft", "--category", "Installation, Getting Started")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{
		"title":        "Setup",
		"slug":         "setup",
		"html_content": "<p>Hi</p>",
		"state":        "draft",
		"category":     []interface{}{"Installation", "Getting Started"},
	}
	if req.Method != "POST" || req.Path != "/api/external/v2/articles" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if !reflect.DeepEqual(req.Body["article"], want) {
		t.Errorf("article = %v, want %v", req.Body["article"], want)
	}
	if out == "" {
		t.Error("expected the created article to be printed")
	}
}

func TestArticlesCreate_OmitsUnsetFields(t *testing.T) {
	req, _, err := runCommand(t, `{"article":{"id":"a1"}}`, "articles", "create", "--title", "Setup", "--category", "Installation")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{"title": "Setup", "category": []interface{}{"Installation"}}
	if !reflect.DeepEqual(req.Body["article"], want) {
		t.Errorf("article = %v, want %v", req.Body["article"], want)
	}
}

func TestArticlesCreate_RejectsUnknownState(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "articles", "create", "--category", "Installation", "--state", "archived")
	if err == nil || err.Error() != `invalid state "archived": use one of draft, published, all` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("no request should be sent, got %s %s", req.Method, req.Path)
	}
}

func TestArticlesUpdate_MovesArticleToCategoryPath(t *testing.T) {
	req, _, err := runCommand(t, `{"article":{"id":"a1"}}`,
		"articles", "update", "a1", "--category", "Installation,Getting Started")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{"category": []interface{}{"Installation", "Getting Started"}}
	if req.Method != "PATCH" || req.Path != "/api/external/v2/articles/a1" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if !reflect.DeepEqual(req.Body["article"], want) {
		t.Errorf("article = %v, want %v", req.Body["article"], want)
	}
}

func TestArticlesUpdate_SendsOnlyChangedFields(t *testing.T) {
	req, _, err := runCommand(t, `{"article":{"id":"a1"}}`,
		"articles", "update", "a1", "--title", "New title", "--state", "published")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{"title": "New title", "state": "published"}
	if !reflect.DeepEqual(req.Body["article"], want) {
		t.Errorf("article = %v, want %v", req.Body["article"], want)
	}
}

func TestArticlesUpdate_RejectsUnknownState(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "articles", "update", "a1", "--state", "archived")
	if err == nil {
		t.Error("expected an error")
	}
	if req.Method != "" {
		t.Errorf("no request should be sent, got %s %s", req.Method, req.Path)
	}
}

func TestArticlesList_SendsStateAndSortParams(t *testing.T) {
	req, _, err := runCommand(t, `{"articles":[{"id":"a1","title":"Setup"}]}`,
		"articles", "list", "--state", "all", "--sort-by", "views_count", "--order-by", "desc")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "GET" || req.Path != "/api/external/v2/articles" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	for key, want := range map[string]string{"state": "all", "sort_by": "views_count", "order_by": "desc"} {
		if got := req.Query.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestArticlesList_OmitsUnsetParams(t *testing.T) {
	req, _, err := runCommand(t, `{"articles":[]}`, "articles", "list")
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"state", "sort_by", "order_by"} {
		if req.Query.Has(key) {
			t.Errorf("%s should not be sent", key)
		}
	}
}

func TestArticlesList_RejectsInvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"state", []string{"--state", "deleted"}, `invalid state "deleted": use one of draft, published, archived, all`},
		{"sort field", []string{"--sort-by", "name"}, `invalid sort field "name": use one of title, slug, state, category_name, views_count, created_at, updated_at`},
		{"order", []string{"--order-by", "up"}, `invalid order "up": use asc or desc`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _, err := runCommand(t, `{}`, append([]string{"articles", "list"}, tt.args...)...)
			if err == nil || err.Error() != tt.want {
				t.Errorf("err = %v, want %s", err, tt.want)
			}
			if req.Method != "" {
				t.Errorf("no request should be sent, got %s %s", req.Method, req.Path)
			}
		})
	}
}
