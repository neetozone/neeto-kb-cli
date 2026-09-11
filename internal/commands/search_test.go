package commands

import (
	"encoding/json"
	"testing"
)

func TestCompactMatches_KeepsOnlyIdSnippetAndURL(t *testing.T) {
	data := json.RawMessage(`{
		"matches": [
			{
				"id": "f58c9e57",
				"title": "How do applications work?",
				"category": "Applications",
				"url": "http://spinkart.lvh.me:8860/articles/how-do-applications-work",
				"category_url": "http://spinkart.lvh.me:8860/folders/applications",
				"matched_title": null,
				"matched_category": null,
				"matched_content": "steps:\nCompilation or <span class=article-matched-tag>Interpretation</span>"
			}
		],
		"pagination": { "total_pages": 3 }
	}`)

	var payload struct {
		Matches    []map[string]interface{} `json:"matches"`
		Pagination map[string]interface{}   `json:"pagination"`
	}
	if err := json.Unmarshal(compactMatches(data), &payload); err != nil {
		t.Fatalf("compactMatches produced invalid JSON: %v", err)
	}

	match := payload.Matches[0]
	if len(match) != 3 {
		t.Fatalf("expected 3 fields, got %d: %v", len(match), match)
	}
	if match["id"] != "f58c9e57" {
		t.Errorf("id = %v", match["id"])
	}
	if match["url"] != "http://spinkart.lvh.me:8860/articles/how-do-applications-work" {
		t.Errorf("url = %v", match["url"])
	}
	if match["matched_content"] != "steps: Compilation or Interpretation" {
		t.Errorf("matched_content = %q", match["matched_content"])
	}
	if payload.Pagination["total_pages"] != float64(3) {
		t.Errorf("pagination was dropped: %v", payload.Pagination)
	}
}

func TestCompactMatches_FallsBackToTitleThenCategory(t *testing.T) {
	tests := []struct {
		name  string
		match string
		want  interface{}
	}{
		{
			name:  "title when content has no match",
			match: `{"matched_content": null, "matched_title": "Billing &amp; <span>refunds</span>"}`,
			want:  "Billing & refunds",
		},
		{
			name:  "category when neither content nor title match",
			match: `{"matched_content": "", "matched_title": null, "matched_category": "<span>Security</span>"}`,
			want:  "Security",
		},
		{
			name:  "nothing matched",
			match: `{"matched_content": null, "matched_title": null, "matched_category": null}`,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := json.RawMessage(`{"matches": [` + tt.match + `]}`)

			var payload struct {
				Matches []map[string]interface{} `json:"matches"`
			}
			if err := json.Unmarshal(compactMatches(data), &payload); err != nil {
				t.Fatalf("compactMatches produced invalid JSON: %v", err)
			}

			if got := payload.Matches[0]["matched_content"]; got != tt.want {
				t.Errorf("matched_content = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompactMatches_LeavesUnexpectedPayloadsAlone(t *testing.T) {
	for _, data := range []json.RawMessage{
		json.RawMessage(`{"errors": ["Search term is required"]}`),
		json.RawMessage(`not json`),
	} {
		if got := string(compactMatches(data)); got != string(data) {
			t.Errorf("compactMatches(%s) = %s, want it untouched", data, got)
		}
	}
}
