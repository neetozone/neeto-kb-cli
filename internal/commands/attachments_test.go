package commands

import (
	"errors"
	"testing"
)

func fakeRead(path string) ([]byte, error) {
	if path == "missing.png" {
		return nil, errors.New("no such file")
	}
	return []byte("hello"), nil
}

func TestBuildAttachmentBody(t *testing.T) {
	t.Run("url", func(t *testing.T) {
		body, err := buildAttachmentBody("https://example.test/a.png", "", "", "", fakeRead)
		if err != nil || len(body) != 1 || body["url"] != "https://example.test/a.png" {
			t.Fatalf("body = %v, err = %v", body, err)
		}
	})

	t.Run("file defaults filename and content type", func(t *testing.T) {
		body, err := buildAttachmentBody("", "dir/diagram.png", "", "", fakeRead)
		if err != nil {
			t.Fatal(err)
		}
		if body["content_base64"] != "aGVsbG8=" || body["filename"] != "diagram.png" || body["content_type"] != "image/png" {
			t.Errorf("body = %v", body)
		}
	})

	t.Run("file honours overrides", func(t *testing.T) {
		body, _ := buildAttachmentBody("", "dir/x.bin", "report.pdf", "application/pdf", fakeRead)
		if body["filename"] != "report.pdf" || body["content_type"] != "application/pdf" {
			t.Errorf("body = %v", body)
		}
	})

	t.Run("neither", func(t *testing.T) {
		if _, err := buildAttachmentBody("", "", "", "", fakeRead); err == nil || err.Error() != "give exactly one of --url or --file" {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("both", func(t *testing.T) {
		if _, err := buildAttachmentBody("https://example.test/a.png", "a.png", "", "", fakeRead); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("unreadable file", func(t *testing.T) {
		if _, err := buildAttachmentBody("", "missing.png", "", "", fakeRead); err == nil {
			t.Error("expected an error")
		}
	})
}
