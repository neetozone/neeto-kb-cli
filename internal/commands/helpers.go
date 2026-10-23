package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/spf13/cobra"
)

func flagValues(cmd *cobra.Command, specs ...string) map[string]interface{} {
	values := map[string]interface{}{}
	for _, spec := range specs {
		name, key, _ := strings.Cut(spec, "=")
		csv := strings.HasSuffix(name, "[]")
		name = strings.TrimSuffix(name, "[]")
		if key == "" {
			key = strings.ReplaceAll(name, "-", "_")
		}

		flag := cmd.Flags().Lookup(name)
		if flag == nil || !flag.Changed {
			continue
		}

		switch {
		case csv:
			list := cli.SplitCSV(flag.Value.String())
			if list == nil {
				list = []string{}
			}
			values[key] = list
		case flag.Value.Type() == "bool":
			values[key], _ = cmd.Flags().GetBool(name)
		case flag.Value.Type() == "int":
			values[key], _ = cmd.Flags().GetInt(name)
		case flag.Value.Type() == "stringArray":
			values[key], _ = cmd.Flags().GetStringArray(name)
		default:
			values[key] = flag.Value.String()
		}
	}
	return values
}

func queryValues(cmd *cobra.Command, params url.Values, specs ...string) url.Values {
	for key, value := range flagValues(cmd, specs...) {
		params.Set(key, fmt.Sprint(value))
	}
	return params
}

func requireFlags(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		_ = cmd.MarkFlagRequired(name)
	}
}

func readJSONFile(path string) (map[string]interface{}, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("%s is not a JSON object: %w", path, err)
	}
	return payload, nil
}

func fetch(cmd *cobra.Command, path string, params url.Values) (json.RawMessage, error) {
	c, err := getClient(cmd)
	if err != nil {
		return nil, err
	}
	return c.Get(path, params)
}

func send(cmd *cobra.Command, method, path string, body any) (json.RawMessage, error) {
	c, err := getClient(cmd)
	if err != nil {
		return nil, err
	}

	switch method {
	case http.MethodPost:
		return c.Post(path, body)
	case http.MethodPut:
		return c.Put(path, body)
	case http.MethodPatch:
		return c.Patch(path, body)
	default:
		return c.DeleteWithBody(path, body)
	}
}

func removeAndConfirm(cmd *cobra.Command, path string, params url.Values, message string) error {
	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	if err := c.DeleteWithParams(path, params); err != nil {
		return err
	}
	printMessage(message)
	return nil
}
