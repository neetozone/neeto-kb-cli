package commands

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestFlagValues(t *testing.T) {
	newCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "x"}
		cmd.Flags().String("name", "", "")
		cmd.Flags().String("parent", "", "")
		cmd.Flags().Bool("private", false, "")
		cmd.Flags().Int("position", 0, "")
		cmd.Flags().String("tags", "", "")
		cmd.Flags().StringArray("event", nil, "")
		return cmd
	}
	specs := []string{"name", "parent=parent_id", "private=is_private", "position", "tags[]", "event=events"}

	tests := []struct {
		name string
		args []string
		want map[string]interface{}
	}{
		{"nothing set", nil, map[string]interface{}{}},
		{"renames keys and types values", []string{"--name", "A", "--parent", "p1", "--private=false", "--position", "2"},
			map[string]interface{}{"name": "A", "parent_id": "p1", "is_private": false, "position": 2}},
		{"sends an explicitly empty string", []string{"--parent", ""}, map[string]interface{}{"parent_id": ""}},
		{"splits comma lists", []string{"--tags", "a, b"}, map[string]interface{}{"tags": []string{"a", "b"}}},
		{"empty comma list clears", []string{"--tags", ""}, map[string]interface{}{"tags": []string{}}},
		{"repeated flag", []string{"--event", "create", "--event", "update"}, map[string]interface{}{"events": []string{"create", "update"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCmd()
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatal(err)
			}
			if got := flagValues(cmd, specs...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadJSONFile(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	bad := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(good, []byte(`{"header_links":[]}`), 0o600)
	_ = os.WriteFile(bad, []byte(`[1]`), 0o600)

	got, err := readJSONFile(good)
	if err != nil || !reflect.DeepEqual(got, map[string]interface{}{"header_links": []interface{}{}}) {
		t.Errorf("got %v, %v", got, err)
	}
	if _, err := readJSONFile(bad); err == nil {
		t.Error("expected an error for a JSON array")
	}
	if _, err := readJSONFile(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestParseColors(t *testing.T) {
	got, err := parseColors([]string{"header-bg=#ffffff", "body-bg=#000"})
	if err != nil || !reflect.DeepEqual(got, map[string]string{"header-bg": "#ffffff", "body-bg": "#000"}) {
		t.Errorf("got %v, %v", got, err)
	}
	for _, bad := range []string{"header-bg", "=#fff", "header-bg="} {
		if _, err := parseColors([]string{bad}); err == nil {
			t.Errorf("expected an error for %q", bad)
		}
	}
}

func TestParseEntities(t *testing.T) {
	got, err := parseEntities([]string{"Article:a1", "Category:c1"})
	want := []map[string]string{{"type": "Article", "id": "a1"}, {"type": "Category", "id": "c1"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, %v", got, err)
	}
	for _, bad := range []string{"a1", "Page:a1", "Article:"} {
		if _, err := parseEntities([]string{bad}); err == nil {
			t.Errorf("expected an error for %q", bad)
		}
	}
}
