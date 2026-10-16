package commands

import "testing"

func TestAuthorsList_SendsSortParams(t *testing.T) {
	req, _, err := runCommand(t, `{"authors":[]}`, "authors", "list", "--sort-by", "articles_count", "--order-by", "asc")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "GET" || req.Path != "/api/external/v2/authors" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if req.Query.Get("sort_by") != "articles_count" || req.Query.Get("order_by") != "asc" {
		t.Errorf("query = %v", req.Query)
	}
}

func TestAuthorsList_OmitsUnsetSortParams(t *testing.T) {
	req, _, err := runCommand(t, `{"authors":[]}`, "authors", "list")
	if err != nil {
		t.Fatal(err)
	}

	if req.Query.Has("sort_by") || req.Query.Has("order_by") {
		t.Errorf("query = %v", req.Query)
	}
}

func TestAuthorsList_RejectsArticleSortFields(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "authors", "list", "--sort-by", "title")
	if err == nil || err.Error() != `invalid sort field "title": use one of name, articles_count` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("no request should be sent, got %s %s", req.Method, req.Path)
	}
}
