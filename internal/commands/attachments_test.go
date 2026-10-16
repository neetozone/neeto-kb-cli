package commands

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAttachmentsCreate_SendsUrl(t *testing.T) {
	req, out, err := runCommand(t, `{"attachment":{"id":"x1"}}`,
		"attachments", "create", "--url", "https://example.test/a.png")
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != "POST" || req.Path != "/api/external/v2/attachments" {
		t.Errorf("request = %s %s", req.Method, req.Path)
	}
	if want := map[string]interface{}{"url": "https://example.test/a.png"}; !reflect.DeepEqual(req.Body, want) {
		t.Errorf("body = %v, want %v", req.Body, want)
	}
	if out == "" {
		t.Error("expected the attachment to be printed")
	}
}

func TestAttachmentsCreate_EncodesFileAndDefaultsNameAndType(t *testing.T) {
	path := writeFile(t, "diagram.png", "hello")

	req, _, err := runCommand(t, `{}`, "attachments", "create", "--file", path)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{
		"content_base64": "aGVsbG8=",
		"filename":       "diagram.png",
		"content_type":   "image/png",
	}
	if !reflect.DeepEqual(req.Body, want) {
		t.Errorf("body = %v, want %v", req.Body, want)
	}
}

func TestAttachmentsCreate_HonoursFilenameAndContentTypeFlags(t *testing.T) {
	path := writeFile(t, "x.bin", "hello")

	req, _, err := runCommand(t, `{}`,
		"attachments", "create", "--file", path, "--filename", "report.pdf", "--content-type", "application/pdf")
	if err != nil {
		t.Fatal(err)
	}

	if req.Body["filename"] != "report.pdf" || req.Body["content_type"] != "application/pdf" {
		t.Errorf("body = %v", req.Body)
	}
}

func TestAttachmentsCreate_FallsBackToOctetStreamForUnknownExtension(t *testing.T) {
	path := writeFile(t, "data.unknownext", "hello")

	req, _, err := runCommand(t, `{}`, "attachments", "create", "--file", path)
	if err != nil {
		t.Fatal(err)
	}

	if req.Body["content_type"] != "application/octet-stream" {
		t.Errorf("content_type = %v", req.Body["content_type"])
	}
}

func TestAttachmentsCreate_AcceptsFileAtTheSizeLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "limit.bin")
	if err := os.WriteFile(path, make([]byte, maxAttachmentBytes), 0o600); err != nil {
		t.Fatal(err)
	}

	req, _, err := runCommand(t, `{}`, "attachments", "create", "--file", path)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != "POST" {
		t.Errorf("method = %q", req.Method)
	}
}

func TestAttachmentsCreate_RejectsFileOverTheSizeLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "big.bin")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(maxAttachmentBytes + 1); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	req, _, err := runCommand(t, `{}`, "attachments", "create", "--file", path)
	if err == nil {
		t.Error("expected an error")
	}
	if req.Method != "" {
		t.Errorf("no request should be sent, got %s %s", req.Method, req.Path)
	}
}

func TestAttachmentsCreate_RejectsMissingFile(t *testing.T) {
	req, _, err := runCommand(t, `{}`, "attachments", "create", "--file", filepath.Join(t.TempDir(), "missing.png"))
	if err == nil {
		t.Error("expected an error")
	}
	if req.Method != "" {
		t.Errorf("no request should be sent, got %s %s", req.Method, req.Path)
	}
}

func TestAttachmentsCreate_RequiresExactlyOneSource(t *testing.T) {
	path := writeFile(t, "a.png", "hello")
	tests := []struct {
		name string
		args []string
	}{
		{"neither", nil},
		{"both", []string{"--url", "https://example.test/a.png", "--file", path}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _, err := runCommand(t, `{}`, append([]string{"attachments", "create"}, tt.args...)...)
			if err == nil || err.Error() != "give exactly one of --url or --file" {
				t.Errorf("err = %v", err)
			}
			if req.Method != "" {
				t.Errorf("no request should be sent, got %s %s", req.Method, req.Path)
			}
		})
	}
}
