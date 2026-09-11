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

func TestPlainText_FlattensWithoutLosingContent(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "comparison operators survive because the API escapes them",
			text: "Set a &lt; b and b &gt; c before you &lt;save&gt; it",
			want: "Set a < b and b > c before you <save> it",
		},
		{
			name: "highlight markup around an escaped operator",
			text: "when x &lt; <span class=article-matched-tag>limit</span>",
			want: "when x < limit",
		},
		{
			name: "multi-byte characters next to markup",
			text: "Réservé 🎉 <span class=article-matched-tag>naïve</span> café",
			want: "Réservé 🎉 naïve café",
		},
		{
			name: "newlines and tabs collapse to single spaces",
			text: "steps:\n\n\tCompilation\r\n  or linking",
			want: "steps: Compilation or linking",
		},
		{
			name: "terminal escape sequences cannot reach the table",
			text: "red \x1b[31malert\x1b[0m and a \x07bell",
			want: "red [31malert [0m and a bell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := plainText(tt.text); got != tt.want {
				t.Errorf("plainText(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestCompactMatches_KeepsEveryPaginationField(t *testing.T) {
	data := json.RawMessage(`{
		"matches": [{ "id": "f58c9e57", "url": "http://example.test/articles/a", "matched_content": "hit" }],
		"pagination": { "total_records": 42, "total_pages": 2, "current_page_number": 1, "page_size": 25 }
	}`)

	var payload struct {
		Pagination map[string]interface{} `json:"pagination"`
	}
	if err := json.Unmarshal(compactMatches(data), &payload); err != nil {
		t.Fatalf("compactMatches produced invalid JSON: %v", err)
	}

	want := map[string]float64{"total_records": 42, "total_pages": 2, "current_page_number": 1, "page_size": 25}
	if len(payload.Pagination) != len(want) {
		t.Fatalf("expected %d pagination fields, got %v", len(want), payload.Pagination)
	}
	for field, value := range want {
		if payload.Pagination[field] != value {
			t.Errorf("pagination[%q] = %v, want %v", field, payload.Pagination[field], value)
		}
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
