package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	return buf.String()
}

func TestUseJSON_ForceJSON(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	if !UseJSON() {
		t.Error("UseJSON() = false, want true when ForceJSON is set")
	}
}

func TestUseJSON_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	if !UseJSON() {
		t.Error("UseJSON() = false, want true when QuietMode is set")
	}
}

func TestPrintMessage_JSON(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	out := captureStdout(t, func() {
		PrintMessage("hello world")
	})

	var parsed map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["message"] != "hello world" {
		t.Errorf("message = %q, want %q", parsed["message"], "hello world")
	}
}

func TestPrintMessage_Plain(t *testing.T) {
	ForceJSON = false
	QuietMode = false

	out := captureStdout(t, func() {
		PrintMessage("hello world")
	})

	trimmed := strings.TrimSpace(out)
	if !strings.Contains(trimmed, "hello world") {
		t.Errorf("output = %q, want it to contain %q", trimmed, "hello world")
	}
}

func TestPrint_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	data := json.RawMessage(`[{"id":1}]`)
	out := captureStdout(t, func() {
		Print(data, nil)
	})

	trimmed := strings.TrimSpace(out)
	if trimmed != `[{"id":1}]` {
		t.Errorf("quiet output = %q, want raw data", trimmed)
	}
}

func TestPrint_JSONEnvelope(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	data := json.RawMessage(`{"name":"test"}`)
	breadcrumbs := []Breadcrumb{{Label: "details", Command: "app show 1"}}

	out := captureStdout(t, func() {
		Print(data, breadcrumbs)
	})

	var envelope Envelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	var parsed map[string]string
	if err := json.Unmarshal(envelope.Data, &parsed); err != nil {
		t.Fatalf("envelope data is not valid JSON: %v", err)
	}
	if parsed["name"] != "test" {
		t.Errorf("data.name = %q, want %q", parsed["name"], "test")
	}
	if len(envelope.Breadcrumbs) != 1 || envelope.Breadcrumbs[0].Label != "details" {
		t.Errorf("breadcrumbs = %v, want [{details app show 1}]", envelope.Breadcrumbs)
	}
}

func TestPrintWithPagination_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	data := json.RawMessage(`[{"id":1}]`)
	pagination := json.RawMessage(`{"current_page_number":1,"total_pages":3,"total_records":25}`)

	out := captureStdout(t, func() {
		PrintWithPagination(data, pagination, nil)
	})

	trimmed := strings.TrimSpace(out)
	if trimmed != `[{"id":1}]` {
		t.Errorf("quiet output = %q, want raw data without pagination", trimmed)
	}
}

func TestPrintPretty_NestedArticleEnvelopeRendersFields(t *testing.T) {
	ForceJSON = false
	QuietMode = false
	ToonMode = false

	longContent := "<p>" + strings.Repeat("lorem ipsum ", 40) + "</p>"
	data := json.RawMessage(`{
		"article": {
			"title": "Getting started with NeetoKB CLI",
			"slug": "getting-started",
			"state": "draft",
			"category": {"id": 5, "name": "Onboarding", "slug": "onboarding"},
			"html_content": ` + strconv.Quote(longContent) + `
		},
		"meta": {"url": "https://example.test", "page_title": "Getting started"}
	}`)

	out := captureStdout(t, func() {
		printPretty(data)
	})

	for _, want := range []string{
		"TITLE", "Getting started with NeetoKB CLI",
		"SLUG", "getting-started",
		"STATE", "draft",
		"CATEGORY", "Onboarding",
		"META",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n--- output ---\n%s", want, out)
		}
	}

	if strings.Contains(out, "fields)") {
		t.Errorf("nested object was collapsed to (N fields) instead of being rendered:\n%s", out)
	}

	if !strings.Contains(out, "...") {
		t.Errorf("long html_content should be truncated to a preview with an ellipsis:\n%s", out)
	}
	if strings.Contains(out, "lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum") {
		t.Errorf("long html_content should not be printed in full:\n%s", out)
	}
}

func TestPickColumns_PriorityThenAlphabeticalScalarsOnly(t *testing.T) {
	sample := map[string]interface{}{
		"title":  "t",
		"slug":   "s",
		"email":  "e",
		"zebra":  "z",
		"apple":  "a",
		"nested": map[string]interface{}{"x": 1},
	}

	got := pickColumns(sample)
	want := []string{"title", "email", "slug", "apple", "zebra"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("pickColumns() = %v, want %v", got, want)
	}
}

func TestPrintPretty_SingleResourceFlattenedRendersScalars(t *testing.T) {
	ForceJSON = false
	QuietMode = false
	ToonMode = false

	data := json.RawMessage(`{"article":{"id":"a-12345678","title":"Updated draft title","slug":"updated","state":"draft"}}`)

	out := captureStdout(t, func() {
		printPretty(data)
	})

	for _, want := range []string{"TITLE", "Updated draft title", "STATE", "draft"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; output:\n%s", want, out)
		}
	}
}

func TestPrintWithPagination_JSONEnvelope(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	data := json.RawMessage(`[{"id":1}]`)
	pagination := json.RawMessage(`{"current_page_number":1,"total_pages":3}`)

	out := captureStdout(t, func() {
		PrintWithPagination(data, pagination, nil)
	})

	var envelope Envelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if envelope.Pagination == nil {
		t.Error("pagination should be present in envelope")
	}
}
