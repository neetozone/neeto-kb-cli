package commands

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

var flagValueSpecs = []string{"name", "parent=parent_id", "private=is_private", "position", "tags[]", "event=events"}

func parsedFlagCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "x"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("parent", "", "")
	cmd.Flags().Bool("private", false, "")
	cmd.Flags().Int("position", 0, "")
	cmd.Flags().String("tags", "", "")
	cmd.Flags().StringArray("event", nil, "")
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestFlagValues_UnsetFlagsAreLeftOut(t *testing.T) {
	got := flagValues(parsedFlagCommand(t), flagValueSpecs...)

	if !reflect.DeepEqual(got, map[string]interface{}{}) {
		t.Errorf("got %v", got)
	}
}

func TestFlagValues_SetFlagsAreRenamedAndKeepTheirTypes(t *testing.T) {
	cmd := parsedFlagCommand(t, "--name", "A", "--parent", "p1", "--private=false", "--position", "2")

	got := flagValues(cmd, flagValueSpecs...)

	want := map[string]interface{}{"name": "A", "parent_id": "p1", "is_private": false, "position": 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFlagValues_ExplicitlyEmptyStringIsSent(t *testing.T) {
	got := flagValues(parsedFlagCommand(t, "--parent", ""), flagValueSpecs...)

	if !reflect.DeepEqual(got, map[string]interface{}{"parent_id": ""}) {
		t.Errorf("got %v", got)
	}
}

func TestFlagValues_CommaListIsSplitAndTrimmed(t *testing.T) {
	got := flagValues(parsedFlagCommand(t, "--tags", "a, b"), flagValueSpecs...)

	if !reflect.DeepEqual(got, map[string]interface{}{"tags": []string{"a", "b"}}) {
		t.Errorf("got %v", got)
	}
}

func TestFlagValues_EmptyCommaListIsSentAsAnEmptyList(t *testing.T) {
	got := flagValues(parsedFlagCommand(t, "--tags", ""), flagValueSpecs...)

	if !reflect.DeepEqual(got, map[string]interface{}{"tags": []string{}}) {
		t.Errorf("got %v", got)
	}
}

func TestFlagValues_RepeatedFlagIsCollected(t *testing.T) {
	got := flagValues(parsedFlagCommand(t, "--event", "create", "--event", "update"), flagValueSpecs...)

	if !reflect.DeepEqual(got, map[string]interface{}{"events": []string{"create", "update"}}) {
		t.Errorf("got %v", got)
	}
}
