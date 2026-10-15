package commands

import (
	"reflect"
	"testing"
)

func TestArticleUpdateBody_CarriesCategoryPath(t *testing.T) {
	body := articleBody("", "", "", "", "Installation, Getting Started")
	article := body["article"].(map[string]interface{})
	want := []string{"Installation", "Getting Started"}
	if !reflect.DeepEqual(article["category"], want) {
		t.Errorf("category = %v, want %v", article["category"], want)
	}
	if len(article) != 1 {
		t.Errorf("unexpected keys: %v", article)
	}
}

func TestArticleUpdateBody_OmitsUnsetFields(t *testing.T) {
	article := articleBody("T", "", "", "draft", "")["article"].(map[string]interface{})
	if len(article) != 2 || article["title"] != "T" || article["state"] != "draft" {
		t.Errorf("article = %v", article)
	}
}

func TestValidateArticleState(t *testing.T) {
	for _, ok := range []string{"", "draft", "published", "all"} {
		if err := validateArticleState(ok); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	err := validateArticleState("archived")
	if err == nil || err.Error() != `invalid state "archived": use one of draft, published, all` {
		t.Errorf("err = %v", err)
	}
}
