package commands

import "testing"

func TestCategoriesList_FlatFlagSendsFlatParam(t *testing.T) {
	req, _, err := runCommand(t, `{"categories":[]}`, "categories", "list", "--flat")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "GET" || req.Path != "/api/external/v2/categories" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if req.Query.Get("flat") != "true" {
		t.Errorf("query = %v", req.Query)
	}
}

func TestCategoriesList_OmitsFlatParamByDefault(t *testing.T) {
	req, _, err := runCommand(t, `{"categories":[]}`, "categories", "list")
	if err != nil {
		t.Fatal(err)
	}

	if req.Query.Has("flat") {
		t.Errorf("query = %v", req.Query)
	}
}
