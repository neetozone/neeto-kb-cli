package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

func parseColors(values []string) (map[string]string, error) {
	colors := map[string]string{}
	for _, value := range values {
		name, color, found := strings.Cut(value, "=")
		if !found || name == "" || color == "" {
			return nil, fmt.Errorf("invalid color %q: use name=value, for example header-bg=#ffffff", value)
		}
		colors[name] = color
	}
	return colors, nil
}

func imageSignedID(cmd *cobra.Command, value string) (string, error) {
	info, err := os.Stat(value)
	if err != nil || info.IsDir() {
		return value, nil
	}

	body, err := buildAttachmentBody("", value, "", "", readAttachmentFile)
	if err != nil {
		return "", err
	}
	data, err := send(cmd, http.MethodPost, "/attachments", body)
	if err != nil {
		return "", err
	}

	var blob struct {
		SignedID string `json:"signed_id"`
	}
	if err := json.Unmarshal(data, &blob); err != nil {
		return "", fmt.Errorf("upload of %s returned an unreadable response: %w", value, err)
	}
	if blob.SignedID == "" {
		return "", fmt.Errorf("upload of %s returned no signed_id: %s", value, data)
	}
	return blob.SignedID, nil
}

func themeConfigurationBody(cmd *cobra.Command) (map[string]interface{}, error) {
	configuration := flagValues(cmd, "title-font-id", "body-font-id", "brand-image-height", "custom-css")

	values, _ := cmd.Flags().GetStringArray("color")
	if len(values) > 0 {
		colors, err := parseColors(values)
		if err != nil {
			return nil, err
		}
		configuration["colors"] = colors
	}

	for _, flag := range []string{"brand-image", "header-image"} {
		value, _ := cmd.Flags().GetString(flag)
		if value == "" {
			continue
		}
		signedID, err := imageSignedID(cmd, value)
		if err != nil {
			return nil, err
		}
		configuration[strings.ReplaceAll(flag, "-", "_")] = signedID
	}
	return map[string]interface{}{"configuration": configuration}, nil
}

var themesCmd = &cobra.Command{
	Use:   "themes",
	Short: "Manage themes",
}

var themesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List built-in and custom themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, "/themes", nil)
		if err != nil {
			return err
		}

		printList(data, "themes", []output.Breadcrumb{
			{Label: "Show", Command: "neetokb themes show <id>"},
			{Label: "Clone", Command: "neetokb themes clone <id>"},
			{Label: "Publish", Command: "neetokb themes publish <id>"},
		})
		return nil
	},
}

var themesShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/themes/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Configuration", Command: "neetokb themes config show <id>"},
		})
		return nil
	},
}

var themesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Rename a custom theme; built-in themes cannot be changed, clone them first",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]interface{}{"theme": flagValues(cmd, "name")}
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/themes/%s", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var themesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a custom theme; the published theme cannot be deleted",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/themes/%s", args[0]), url.Values{}, "Theme deleted.")
	},
}

var themesCloneCmd = &cobra.Command{
	Use:   "clone <id>",
	Short: "Copy a built-in or custom theme into a new unpublished custom theme that can be edited",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, fmt.Sprintf("/themes/%s/clone", args[0]), nil)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Edit", Command: "neetokb themes config update <id>"},
			{Label: "Publish", Command: "neetokb themes publish <id>"},
		})
		return nil
	},
}

var themesPublishCmd = &cobra.Command{
	Use:   "publish <id>",
	Short: "Make a theme the live theme of the public knowledge base",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := send(cmd, http.MethodPost, fmt.Sprintf("/themes/%s/publish", args[0]), nil)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var themesConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the fonts, colors, images and custom CSS of a theme",
}

var themesConfigShowCmd = &cobra.Command{
	Use:   "show <theme-id>",
	Short: "Show the configuration of a theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := fetch(cmd, fmt.Sprintf("/themes/%s/configuration", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetokb themes config update <theme-id>"},
		})
		return nil
	},
}

var themesConfigUpdateCmd = &cobra.Command{
	Use:   "update <theme-id>",
	Short: "Change the configuration of a custom theme; colors are merged into the current ones",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := themeConfigurationBody(cmd)
		if err != nil {
			return err
		}

		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/themes/%s/configuration", args[0]), body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var themesFontsCmd = &cobra.Command{
	Use:   "fonts",
	Short: "Manage the fonts of a custom theme",
}

var themesFontsCreateCmd = &cobra.Command{
	Use:   "create <theme-id>",
	Short: "Add a Google font to a custom theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]interface{}{"font": flagValues(cmd, "name", "weights")}
		data, err := send(cmd, http.MethodPost, fmt.Sprintf("/themes/%s/fonts", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var themesFontsUpdateCmd = &cobra.Command{
	Use:   "update <theme-id> <font-id>",
	Short: "Change a font of a custom theme",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]interface{}{"font": flagValues(cmd, "name", "weights")}
		data, err := send(cmd, http.MethodPatch, fmt.Sprintf("/themes/%s/fonts/%s", args[0], args[1]), body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var themesFontsDeleteCmd = &cobra.Command{
	Use:   "delete <theme-id> <font-id>",
	Short: "Remove a font from a custom theme",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAndConfirm(cmd, fmt.Sprintf("/themes/%s/fonts/%s", args[0], args[1]), url.Values{}, "Font deleted.")
	},
}

func init() {
	themesUpdateCmd.Flags().String("name", "", "New theme name")
	requireFlags(themesUpdateCmd, "name")

	themesConfigUpdateCmd.Flags().String("title-font-id", "", "ID of a font of this theme for titles")
	themesConfigUpdateCmd.Flags().String("body-font-id", "", "ID of a font of this theme for body text")
	themesConfigUpdateCmd.Flags().Int("brand-image-height", 0, "Brand image height in pixels")
	themesConfigUpdateCmd.Flags().String("custom-css", "", "Custom CSS; an empty value clears it")
	themesConfigUpdateCmd.Flags().StringArray("color", nil, "Color as name=value, for example header-bg=#ffffff; repeat for several")
	themesConfigUpdateCmd.Flags().String("brand-image", "", "Brand image: a local file path or the signed ID of an uploaded attachment")
	themesConfigUpdateCmd.Flags().String("header-image", "", "Header image: a local file path or the signed ID of an uploaded attachment")

	themesFontsCreateCmd.Flags().String("name", "", "Google font name, for example Roboto")
	themesFontsCreateCmd.Flags().String("weights", "", "Font weights (comma-separated), for example 400,700")
	requireFlags(themesFontsCreateCmd, "name")
	themesFontsUpdateCmd.Flags().String("name", "", "Google font name")
	themesFontsUpdateCmd.Flags().String("weights", "", "Font weights (comma-separated)")

	themesFontsCmd.AddCommand(themesFontsCreateCmd, themesFontsUpdateCmd, themesFontsDeleteCmd)
	themesConfigCmd.AddCommand(themesConfigShowCmd, themesConfigUpdateCmd)
	themesCmd.AddCommand(themesListCmd, themesShowCmd, themesUpdateCmd, themesDeleteCmd, themesCloneCmd, themesPublishCmd, themesConfigCmd, themesFontsCmd)
	register(func(root *cobra.Command) { root.AddCommand(themesCmd) })
}
