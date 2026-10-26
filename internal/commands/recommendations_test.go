package commands

import (
	"reflect"
	"testing"
)

func TestRecommendationsCreate_EntitiesAreSentAsTypedReferences(t *testing.T) {
	req, _, err := runCommand(t, `{"recommendation":{"id":"r1"}}`,
		"recommendations", "create", "--title", "Billing", "--url", "/billing",
		"--entity", "Article:a1", "--entity", "Category:c1")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "POST" || req.Path != "/api/external/v2/recommendations" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := []interface{}{
		map[string]interface{}{"type": "Article", "id": "a1"},
		map[string]interface{}{"type": "Category", "id": "c1"},
	}
	if !reflect.DeepEqual(req.Body["entities"], want) {
		t.Errorf("entities = %v", req.Body["entities"])
	}
}

func TestRecommendationsCreate_EntityWithoutTypeIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "recommendations", "create", "--title", "Billing", "--url", "/billing", "--entity", "a1")

	if err == nil || err.Error() != `invalid entity "a1": use Article:<id> or Category:<id>` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}

func TestRecommendationsCreate_EntityWithoutIDIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "recommendations", "create", "--title", "Billing", "--url", "/billing", "--entity", "Article:")

	if err == nil || err.Error() != `invalid entity "Article:": use Article:<id> or Category:<id>` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}

func TestRecommendationsCreate_UnknownEntityTypeIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "recommendations", "create", "--title", "Billing", "--url", "/billing", "--entity", "Page:p1")

	if err == nil || err.Error() != `invalid entity "Page:p1": use Article:<id> or Category:<id>` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}

func TestRecommendationsUpdate_ClearEntitiesSendsAnEmptyList(t *testing.T) {
	req, _, err := runCommand(t, `{"recommendation":{"id":"r1"}}`, "recommendations", "update", "r1", "--clear-entities")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "PATCH" || req.Path != "/api/external/v2/recommendations/r1" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if !reflect.DeepEqual(req.Body["entities"], []interface{}{}) {
		t.Errorf("entities = %v", req.Body["entities"])
	}
}

func TestRecommendationsUpdate_EntityWithClearEntitiesIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "recommendations", "update", "r1", "--entity", "Article:a1", "--clear-entities")

	if err == nil {
		t.Error("expected an error")
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}
