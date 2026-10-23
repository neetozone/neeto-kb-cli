package commands

import (
	"reflect"
	"testing"
)

func TestThemesConfigUpdate_ColorsAreSentUnderConfiguration(t *testing.T) {
	req, _, err := runCommand(t, `{"configuration":{"id":"th1"}}`,
		"themes", "config", "update", "th1", "--color", "header-bg=#ffffff", "--color", "body-bg=#000")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "PATCH" || req.Path != "/api/external/v2/themes/th1/configuration" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	want := map[string]interface{}{"colors": map[string]interface{}{"header-bg": "#ffffff", "body-bg": "#000"}}
	if !reflect.DeepEqual(req.Body["configuration"], want) {
		t.Errorf("configuration = %v", req.Body["configuration"])
	}
}

func TestThemesConfigUpdate_ColorWithoutEqualsSignIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "themes", "config", "update", "th1", "--color", "header-bg")

	if err == nil || err.Error() != `invalid color "header-bg": use name=value, for example header-bg=#ffffff` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}

func TestThemesConfigUpdate_ColorWithoutNameIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "themes", "config", "update", "th1", "--color", "=#fff")

	if err == nil || err.Error() != `invalid color "=#fff": use name=value, for example header-bg=#ffffff` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}

func TestThemesConfigUpdate_ColorWithoutValueIsRefusedBeforeSending(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "themes", "config", "update", "th1", "--color", "header-bg=")

	if err == nil || err.Error() != `invalid color "header-bg=": use name=value, for example header-bg=#ffffff` {
		t.Errorf("err = %v", err)
	}
	if req.Method != "" {
		t.Errorf("request sent: %s %s", req.Method, req.Path)
	}
}
