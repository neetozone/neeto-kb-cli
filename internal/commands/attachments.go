package commands

import (
	"encoding/base64"
	"errors"
	"mime"
	"os"
	"path/filepath"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var attachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "Manage attachments",
}

func buildAttachmentBody(url, path, filename, contentType string, readFile func(string) ([]byte, error)) (map[string]interface{}, error) {
	if (url == "") == (path == "") {
		return nil, errors.New("give exactly one of --url or --file")
	}
	if url != "" {
		return map[string]interface{}{"url": url}, nil
	}

	content, err := readFile(path)
	if err != nil {
		return nil, err
	}
	if filename == "" {
		filename = filepath.Base(path)
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(path))
	}

	return map[string]interface{}{
		"content_base64": base64.StdEncoding.EncodeToString(content),
		"filename":       filename,
		"content_type":   contentType,
	}, nil
}

var attachmentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Upload an attachment from a URL or a local file",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		url, _ := cmd.Flags().GetString("url")
		path, _ := cmd.Flags().GetString("file")
		filename, _ := cmd.Flags().GetString("filename")
		contentType, _ := cmd.Flags().GetString("content-type")

		body, err := buildAttachmentBody(url, path, filename, contentType, os.ReadFile)
		if err != nil {
			return err
		}

		data, err := c.Post("/attachments", body)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Use in article", Command: "neetokb articles update <id> --html-content '...<attachable_html>...'"},
		})
		return nil
	},
}

func init() {
	attachmentsCreateCmd.Flags().String("url", "", "Public URL of the file to attach")
	attachmentsCreateCmd.Flags().String("file", "", "Path of a local file to upload")
	attachmentsCreateCmd.Flags().String("filename", "", "File name to store (default: the base name of --file)")
	attachmentsCreateCmd.Flags().String("content-type", "", "MIME type (default: guessed from the --file extension)")

	attachmentsCmd.AddCommand(attachmentsCreateCmd)
	register(func(root *cobra.Command) { root.AddCommand(attachmentsCmd) })
}
