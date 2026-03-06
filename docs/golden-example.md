# Golden Example: `a6 route list` Implementation

This document provides a complete, working implementation of `a6 route list` — the canonical reference. When adding any new command, follow this exact structure. Every file shown here is production-ready and tested.

## File 1: `pkg/cmd/factory.go` — Factory Pattern

The Factory decouples command implementation from global state and dependencies, allowing for easy testing.

```go
package cmd

import (
	"net/http"

	"github.com/api7/a6/pkg/config"
	"github.com/api7/a6/pkg/iostreams"
)

type Factory struct {
	IOStreams *iostreams.IOStreams

	// HttpClient returns a lazy-initialized, auth-injected HTTP client.
	HttpClient func() (*http.Client, error)

	// Config returns a lazy-initialized configuration reader.
	Config func() (config.Config, error)
}
```

## File 2: `pkg/iostreams/iostreams.go` — I/O Abstraction

Abstracts terminal I/O for consistency across real execution and tests.

```go
package iostreams

import (
	"bytes"
	"io"
	"os"
)

type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	inTTY  bool
	outTTY bool
	errTTY bool
}

// System creates IOStreams using real os.Stdin, os.Stdout, and os.Stderr.
func System() *IOStreams {
	return &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		inTTY:  isTerminal(os.Stdin),
		outTTY: isTerminal(os.Stdout),
		errTTY: isTerminal(os.Stderr),
	}
}

// Test creates IOStreams with bytes.Buffer for testing.
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}
	return &IOStreams{
		In:     io.NopCloser(in),
		Out:    out,
		ErrOut: err,
	}, in, out, err
}

func (s *IOStreams) IsStdoutTTY() bool {
	return s.outTTY
}

func (s *IOStreams) ColorEnabled() bool {
	return os.Getenv("NO_COLOR") == "" && s.outTTY
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
```

## File 3: `pkg/api/client.go` — API Client

A generic API client handling authentication, request helpers, and response parsing.

```go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

type apiKeyTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-API-KEY", t.apiKey)
	return t.base.RoundTrip(req)
}

// Generic response types
type ListResponse[T any] struct {
	Total int `json:"total"`
	List  []struct {
		Value T `json:"value"`
	} `json:"list"`
}

type APIError struct {
	StatusCode int
	ErrorMsg   string `json:"error_msg"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status %d, message: %s", e.StatusCode, e.ErrorMsg)
}

func (c *Client) Get(path string, query map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		apiErr.StatusCode = resp.StatusCode
		_ = json.Unmarshal(body, &apiErr)
		return nil, &apiErr
	}

	return body, nil
}
```

## File 4: `pkg/api/types_route.go` — Route Types

Matching the APISIX Admin API route schema.

```go
package api

type Route struct {
	ID         *string                `json:"id,omitempty"`
	URI        *string                `json:"uri,omitempty"`
	URIs       []string               `json:"uris,omitempty"`
	Name       *string                `json:"name,omitempty"`
	Desc       *string                `json:"desc,omitempty"`
	Methods    []string               `json:"methods,omitempty"`
	Host       *string                `json:"host,omitempty"`
	Hosts      []string               `json:"hosts,omitempty"`
	Priority   *int                   `json:"priority,omitempty"`
	Status     *int                   `json:"status,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	Upstream   *Upstream              `json:"upstream,omitempty"`
	UpstreamID *string                `json:"upstream_id,omitempty"`
	ServiceID  *string                `json:"service_id,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	CreateTime *int64                 `json:"create_time,omitempty"`
	UpdateTime *int64                 `json:"update_time,omitempty"`
}

type Upstream struct {
	Type  string                 `json:"type"`
	Nodes map[string]interface{} `json:"nodes"`
}
```

## File 5: `pkg/cmd/route/route.go` — Route Parent Command

```go
package route

import (
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/route/list"
	"github.com/spf13/cobra"
)

func NewCmdRoute(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route <command>",
		Short: "Manage APISIX routes",
		Long:  "Commands for creating, listing, updating, and deleting APISIX routes.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	// Add other subcommands here: get, create, update, delete
	return cmd
}
```

## File 6: `pkg/cmd/route/list/list.go` — Route List Command (THE GOLDEN EXAMPLE)

This is the standard pattern for all list commands.

```go
package list

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ListOptions struct {
	IO     *iostreams.IOStreams
	Client func() (*http.Client, error)
	BaseURL func() (string, error)

	Page     int
	PageSize int
	Name     string
	Label    string
	URI      string
	Output   string
}

func NewCmdList(f *cmd.Factory) *cobra.Command {
	opts := &ListOptions{
		IO: f.IOStreams,
		Client: f.HttpClient,
		BaseURL: func() (string, error) {
			cfg, err := f.Config()
			if err != nil {
				return "", err
			}
			return cfg.BaseURL(), nil
		},
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List APISIX routes",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.PageSize < 1 || opts.PageSize > 500 {
				return fmt.Errorf("page-size must be between 1 and 500")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRun(opts)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&opts.Label, "label", "", "Filter by label (key=value)")
	cmd.Flags().StringVar(&opts.URI, "uri", "", "Filter by URI")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format (json, yaml, table)")

	return cmd
}

func listRun(opts *ListOptions) error {
	httpClient, err := opts.Client()
	if err != nil {
		return err
	}
	baseURL, err := opts.BaseURL()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, baseURL)
	
	queryParams := map[string]string{
		"page":      fmt.Sprintf("%d", opts.Page),
		"page_size": fmt.Sprintf("%d", opts.PageSize),
	}
	if opts.Name != "" {
		queryParams["name"] = opts.Name
	}
	if opts.Label != "" {
		queryParams["label"] = opts.Label
	}
	if opts.URI != "" {
		queryParams["uri"] = opts.URI
	}

	data, err := client.Get("/apisix/admin/routes", queryParams)
	if err != nil {
		return err
	}

	var resp api.ListResponse[api.Route]
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	routes := make([]api.Route, len(resp.List))
	for i, item := range resp.List {
		routes[i] = item.Value
	}

	// Output logic: prioritize --output flag, then detect TTY
	format := opts.Output
	if format == "" {
		if opts.IO.IsStdoutTTY() {
			format = "table"
		} else {
			format = "json"
		}
	}

	switch format {
	case "table":
		return printTable(opts.IO, routes)
	case "json":
		return printJSON(opts.IO, routes)
	case "yaml":
		return printYAML(opts.IO, routes)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func printTable(io *iostreams.IOStreams, routes []api.Route) error {
	if len(routes) == 0 {
		fmt.Fprintln(io.Out, "No routes found.")
		return nil
	}

	w := tabwriter.NewWriter(io.Out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tURI\tMETHODS\tSTATUS\tUPSTREAM")
	
	for _, r := range routes {
		id := "N/A"
		if r.ID != nil { id = *r.ID }
		name := "N/A"
		if r.Name != nil { name = *r.Name }
		uri := "N/A"
		if r.URI != nil { 
			uri = *r.URI 
		} else if len(r.URIs) > 0 {
			uri = strings.Join(r.URIs, ",")
		}
		
		methods := "*"
		if len(r.Methods) > 0 {
			methods = strings.Join(r.Methods, ",")
		}
		
		status := "1"
		if r.Status != nil {
			status = fmt.Sprintf("%d", *r.Status)
		}

		upstream := "N/A"
		if r.UpstreamID != nil {
			upstream = *r.UpstreamID
		} else if r.Upstream != nil {
			upstream = "(embedded)"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, name, uri, methods, status, upstream)
	}
	return w.Flush()
}

func printJSON(io *iostreams.IOStreams, routes []api.Route) error {
	enc := json.NewEncoder(io.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(routes)
}

func printYAML(io *iostreams.IOStreams, routes []api.Route) error {
	return yaml.NewEncoder(io.Out).Encode(routes)
}
```

## File 7: `pkg/cmd/route/list/list_test.go` — Tests

```go
package list

import (
	"net/http"
	"testing"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/httpmock"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestRouteList_TTY(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse("../../test/fixtures/route_list.json"))

	io, _, out, _ := iostreams.Test()
	io.SetStdoutTTY(true) // Force TTY for table output

	f := &cmd.Factory{
		IOStreams: io,
		HttpClient: func() (*http.Client, error) {
			return reg.GetClient(), nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	cmd := NewCmdList(f)
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, out.String(), "ID   NAME        URI")
	assert.Contains(t, out.String(), "1    users-api   /api/v1/users")
	reg.Verify(t)
}

func TestRouteList_NonTTY(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse("../../test/fixtures/route_list.json"))

	io, _, out, _ := iostreams.Test()
	io.SetStdoutTTY(false) // Non-TTY should default to JSON

	f := &cmd.Factory{
		IOStreams: io,
		HttpClient: func() (*http.Client, error) {
			return reg.GetClient(), nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	cmd := NewCmdList(f)
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, out.String(), "\"id\": \"1\"") // Validates JSON output
}

func TestRouteList_APIError(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.StringResponse(403, `{"error_msg":"forbidden"}`))

	io, _, _, _ := iostreams.Test()
	f := &cmd.Factory{
		IOStreams: io,
		HttpClient: func() (*http.Client, error) {
			return reg.GetClient(), nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	cmd := NewCmdList(f)
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}
```

## File 8: `pkg/httpmock/httpmock.go` — HTTP Mock for Tests

```go
package httpmock

import (
	"io"
	"net/http"
	"os"
	"testing"
)

type Response struct {
	StatusCode int
	Body       []byte
}

type Registry struct {
	mocks []struct {
		method string
		path   string
		resp   Response
		called bool
	}
}

func (r *Registry) Register(method, path string, resp Response) {
	r.mocks = append(r.mocks, struct {
		method string
		path   string
		resp   Response
		called bool
	}{method, path, resp, false})
}

func (r *Registry) RoundTrip(req *http.Request) (*http.Response, error) {
	for i, m := range r.mocks {
		if m.method == req.Method && m.path == req.URL.Path {
			r.mocks[i].called = true
			return &http.Response{
				StatusCode: m.resp.StatusCode,
				Body:       io.NopCloser(bytes.NewBuffer(m.resp.Body)),
				Header:     make(http.Header),
			}, nil
		}
	}
	return nil, fmt.Errorf("no mock registered for %s %s", req.Method, req.URL.Path)
}

func (r *Registry) GetClient() *http.Client {
	return &http.Client{Transport: r}
}

func (r *Registry) Verify(t *testing.T) {
	for _, m := range r.mocks {
		if !m.called {
			t.Errorf("mock never called: %s %s", m.method, m.path)
		}
	}
}

func JSONResponse(path string) Response {
	b, _ := os.ReadFile(path)
	return Response{StatusCode: 200, Body: b}
}
```

## File 9: `test/fixtures/route_list.json` — Test Fixture

```json
{
  "total": 2,
  "list": [
    {
      "key": "/apisix/routes/1",
      "value": {
        "id": "1",
        "uri": "/api/v1/users",
        "name": "users-api",
        "methods": ["GET", "POST"],
        "status": 1,
        "upstream": {
          "type": "roundrobin",
          "nodes": {"httpbin.org:80": 1}
        }
      },
      "createdIndex": 1,
      "modifiedIndex": 2
    }
  ]
}
```

## "How to Add a New Command" Checklist

1. Create `pkg/api/types_<resource>.go` with Go structs matching API schema.
2. Create `pkg/cmd/<resource>/<resource>.go` parent command.
3. Create `pkg/cmd/<resource>/<action>/<action>.go` following the Options+NewCmd+Run pattern.
4. Create `pkg/cmd/<resource>/<action>/<action>_test.go` with TTY/non-TTY/error test cases.
5. Add test fixture JSON in `test/fixtures/`.
6. Register in `pkg/cmd/root/root.go`.
7. Update `docs/user-guide/` with user-facing documentation.
8. Run `make check` to verify.
