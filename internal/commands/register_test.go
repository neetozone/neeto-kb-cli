package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	product "github.com/neetozone/neeto-kb-cli"
)

type recordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   map[string]interface{}
}

var buildApp sync.Once

func resetFlags(cmd *cobra.Command) {
	reset := func(flag *pflag.Flag) {
		if flag.Changed {
			_ = flag.Value.Set(flag.DefValue)
			flag.Changed = false
		}
	}
	cmd.Flags().VisitAll(reset)
	cmd.PersistentFlags().VisitAll(reset)
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

func runCommand(t *testing.T, response string, args ...string) (recordedRequest, string, error) {
	t.Helper()

	buildApp.Do(func() {
		cfg, err := config.Parse(product.ConfigYAML)
		if err != nil {
			t.Fatal(err)
		}
		Register(cli.New(*cfg))
	})

	var recorded recordedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		recorded = recordedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query()}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &recorded.Body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("NEETOKB_BASE_URL", server.URL)
	configDir, err := app.Auth.ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	credentials := `{"credentials":[{"subdomain":"demo","email":"oliver.smith@example.test","session_token":"token"}]}`
	if err := os.WriteFile(filepath.Join(configDir, "auth.json"), []byte(credentials), 0o600); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app.Printer.Out = out
	app.Printer.Err = io.Discard
	app.Root().SetOut(io.Discard)
	app.Root().SetErr(io.Discard)
	app.Root().SetArgs(append([]string{"--subdomain", "demo", "--quiet"}, args...))
	t.Cleanup(func() { resetFlags(app.Root()) })

	err = app.Root().Execute()
	return recorded, out.String(), err
}
