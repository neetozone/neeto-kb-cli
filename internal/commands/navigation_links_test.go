package commands

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNavigationLinksUpdate_JSONFileIsPutAsTheBody(t *testing.T) {
	path := filepath.Join(t.TempDir(), "links.json")
	links := `{"header_links":[{"label":"Docs","link":"https://example.com","kind":"external_link"}],"footer_links":[]}`
	if err := os.WriteFile(path, []byte(links), 0o600); err != nil {
		t.Fatal(err)
	}

	req, _, err := runCommand(t, `{"header_links":[],"footer_links":[]}`, "navigation-links", "update", "--json-file", path)
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "PUT" || req.Path != "/api/external/v2/navigation_links" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := map[string]interface{}{
		"header_links": []interface{}{
			map[string]interface{}{"label": "Docs", "link": "https://example.com", "kind": "external_link"},
		},
		"footer_links": []interface{}{},
	}
	if !reflect.DeepEqual(req.Body, want) {
		t.Errorf("body = %v", req.Body)
	}
}
