package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/api7/a6/pkg/selector"
	"github.com/api7/a6/pkg/tableprinter"
)

type Options struct {
	IO            *iostreams.IOStreams
	Config        func() (config.Config, error)
	Client        func() (*http.Client, error)
	ControlClient func() (*http.Client, error)

	ID         string
	Method     string
	Path       string
	Headers    []string
	Body       string
	Host       string
	ControlURL string
	GatewayURL string
	Output     string
}

type SchemaResponse struct {
	Plugins map[string]PluginSchema `json:"plugins"`
}

type PluginSchema struct {
	Priority int `json:"priority"`
}

type TraceResult struct {
	Route             RouteInfo    `json:"route"`
	Request           RequestInfo  `json:"request"`
	Response          ResponseInfo `json:"response"`
	ConfiguredPlugins []PluginInfo `json:"configured_plugins"`
}

type RouteInfo struct {
	ID       string                 `json:"id"`
	URI      string                 `json:"uri"`
	Methods  []string               `json:"methods,omitempty"`
	Hosts    []string               `json:"hosts,omitempty"`
	Upstream interface{}            `json:"upstream,omitempty"`
	Plugins  map[string]interface{} `json:"plugins,omitempty"`
}

type RequestInfo struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type ResponseInfo struct {
	Status          int      `json:"status"`
	StatusText      string   `json:"status_text"`
	DurationMs      int64    `json:"duration_ms"`
	UpstreamStatus  string   `json:"upstream_status,omitempty"`
	ExecutedPlugins []string `json:"executed_plugins,omitempty"`
}

type PluginInfo struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

func NewCmdTrace(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
		Client: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "trace [route-id]",
		Short: "Trace a request through a route",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ID = args[0]
			}
			return traceRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Method, "method", "", "HTTP method for the test request")
	cmd.Flags().StringVar(&opts.Path, "path", "", "Request path for the test request")
	cmd.Flags().StringArrayVar(&opts.Headers, "header", nil, "Request header in 'Key: Value' format (repeatable)")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Request body for the test request")
	cmd.Flags().StringVar(&opts.Host, "host", "", "Host header for the test request")
	cmd.Flags().StringVar(&opts.ControlURL, "control-url", "", "APISIX Control API URL")
	cmd.Flags().StringVar(&opts.GatewayURL, "gateway-url", "", "APISIX Gateway URL")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: table, json, yaml")

	return cmd
}

func traceRun(opts *Options) error {
	if opts.ID == "" {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("route-id argument is required (or run interactively in a terminal)")
		}
		id, err := selectRoute(opts)
		if err != nil {
			return err
		}
		opts.ID = id
	}

	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	adminClient := api.NewClient(httpClient, cfg.BaseURL())
	body, err := adminClient.Get(fmt.Sprintf("/apisix/admin/routes/%s", opts.ID), nil)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var routeResp api.SingleResponse[api.Route]
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	gatewayURL := opts.GatewayURL
	if gatewayURL == "" {
		gatewayURL = os.Getenv("APISIX_GATEWAY_URL")
	}
	if gatewayURL == "" {
		gatewayURL, err = deriveGatewayURL(cfg.BaseURL())
		if err != nil {
			return fmt.Errorf("failed to derive gateway URL: %w", err)
		}
	}

	controlURL := opts.ControlURL
	if controlURL == "" {
		controlURL, err = deriveControlURL(cfg.BaseURL())
		if err != nil {
			return fmt.Errorf("failed to derive control API URL: %w", err)
		}
	}

	pluginPriorities := fetchPluginPriorities(opts, controlURL)

	route := routeResp.Value
	method := resolveMethod(opts.Method, route.Methods)
	path := resolvePath(opts.Path, route.URI, route.URIs)
	host := resolveHost(opts.Host, route.Host, route.Hosts)

	headers, err := parseHeaders(opts.Headers)
	if err != nil {
		return err
	}

	requestURL, err := joinURLPath(gatewayURL, path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, requestURL, bytes.NewBufferString(opts.Body))
	if err != nil {
		return fmt.Errorf("failed to create gateway request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if host != "" {
		req.Host = host
	}

	gatewayClient := &http.Client{Timeout: 15 * time.Second}
	start := time.Now()
	resp, err := gatewayClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send test request to gateway: %w", err)
	}
	durationMs := time.Since(start).Milliseconds()
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	executedPlugins := parsePluginHeader(resp.Header.Get("Apisix-Plugins"))
	configuredPlugins := buildConfiguredPlugins(route.Plugins, pluginPriorities)

	result := TraceResult{
		Route: RouteInfo{
			ID:       routeID(routeResp),
			URI:      routeURI(route.URI, route.URIs, path),
			Methods:  route.Methods,
			Hosts:    allHosts(route.Host, route.Hosts),
			Upstream: routeUpstream(route.Upstream, route.UpstreamID),
			Plugins:  route.Plugins,
		},
		Request: RequestInfo{
			Method:  method,
			URL:     requestURL,
			Headers: resultHeaders(headers, host),
		},
		Response: ResponseInfo{
			Status:          resp.StatusCode,
			StatusText:      http.StatusText(resp.StatusCode),
			DurationMs:      durationMs,
			UpstreamStatus:  resp.Header.Get("X-APISIX-Upstream-Status"),
			ExecutedPlugins: executedPlugins,
		},
		ConfiguredPlugins: configuredPlugins,
	}

	format := opts.Output
	if format == "" {
		if opts.IO.IsStdoutTTY() {
			format = "table"
		} else {
			format = "json"
		}
	}

	if format == "table" {
		printTraceTable(opts.IO.Out, result)
		return nil
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(result)
}

func resolveMethod(flagMethod string, routeMethods []string) string {
	if flagMethod != "" {
		return strings.ToUpper(flagMethod)
	}
	if len(routeMethods) > 0 {
		return strings.ToUpper(routeMethods[0])
	}
	return http.MethodGet
}

func resolvePath(flagPath string, routeURI *string, routeURIs []string) string {
	path := flagPath
	if path == "" && routeURI != nil {
		path = *routeURI
	}
	if path == "" && len(routeURIs) > 0 {
		path = routeURIs[0]
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func resolveHost(flagHost string, routeHost *string, routeHosts []string) string {
	if flagHost != "" {
		return flagHost
	}
	if len(routeHosts) > 0 {
		return routeHosts[0]
	}
	if routeHost != nil {
		return *routeHost
	}
	return ""
}

func parseHeaders(headerFlags []string) (map[string]string, error) {
	if len(headerFlags) == 0 {
		return nil, nil
	}
	headers := make(map[string]string, len(headerFlags))
	for _, h := range headerFlags {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --header value %q, expected 'Key: Value'", h)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("invalid --header value %q, header key is empty", h)
		}
		headers[key] = value
	}
	return headers, nil
}

func joinURLPath(baseURL, path string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid gateway URL: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid gateway URL: missing host")
	}
	u.Path = path
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func deriveGatewayURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("invalid base URL: %s", baseURL)
	}
	return "http://" + net.JoinHostPort(host, "9080"), nil
}

func deriveControlURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("invalid base URL: %s", baseURL)
	}
	return "http://" + net.JoinHostPort(host, "9090"), nil
}

func fetchPluginPriorities(opts *Options, controlURL string) map[string]int {
	client := &http.Client{Timeout: 5 * time.Second}
	if opts.ControlClient != nil {
		customClient, err := opts.ControlClient()
		if err == nil {
			client = customClient
		}
	}

	schemaURL, err := joinURLPath(controlURL, "/v1/schema")
	if err != nil {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, schemaURL, nil)
	if err != nil {
		return nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var schemaResp SchemaResponse
	if err := json.Unmarshal(body, &schemaResp); err != nil {
		return nil
	}

	priorities := make(map[string]int, len(schemaResp.Plugins))
	for name, plugin := range schemaResp.Plugins {
		priorities[name] = plugin.Priority
	}
	return priorities
}

func buildConfiguredPlugins(plugins map[string]interface{}, priorities map[string]int) []PluginInfo {
	if len(plugins) == 0 {
		return nil
	}

	out := make([]PluginInfo, 0, len(plugins))
	for name := range plugins {
		out = append(out, PluginInfo{Name: name, Priority: priorities[name]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].Name < out[j].Name
		}
		return out[i].Priority > out[j].Priority
	})
	return out
}

func parsePluginHeader(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func routeID(resp api.SingleResponse[api.Route]) string {
	if resp.Value.ID != nil {
		return *resp.Value.ID
	}
	return strings.TrimPrefix(resp.Key, "/apisix/routes/")
}

func routeURI(uri *string, uris []string, fallback string) string {
	if uri != nil && *uri != "" {
		return *uri
	}
	if len(uris) > 0 {
		return uris[0]
	}
	return fallback
}

func allHosts(host *string, hosts []string) []string {
	if len(hosts) > 0 {
		return hosts
	}
	if host != nil && *host != "" {
		return []string{*host}
	}
	return nil
}

func routeUpstream(upstream *api.RouteUpstream, upstreamID *string) interface{} {
	if upstream != nil {
		return upstream
	}
	if upstreamID != nil && *upstreamID != "" {
		return map[string]string{"upstream_id": *upstreamID}
	}
	return nil
}

func resultHeaders(headers map[string]string, host string) map[string]string {
	if len(headers) == 0 && host == "" {
		return nil
	}
	out := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		out[k] = v
	}
	if host != "" {
		out["Host"] = host
	}
	return out
}

func printTraceTable(out io.Writer, result TraceResult) {
	fmt.Fprintf(out, "Route:    %s (ID: %s)\n", result.Route.URI, result.Route.ID)
	if len(result.Route.Methods) > 0 {
		fmt.Fprintf(out, "Methods:  %s\n", strings.Join(result.Route.Methods, ", "))
	}
	if len(result.Route.Hosts) > 0 {
		fmt.Fprintf(out, "Hosts:    %s\n", strings.Join(result.Route.Hosts, ", "))
	}
	fmt.Fprintf(out, "Upstream: %s\n\n", formatUpstream(result.Route.Upstream))

	fmt.Fprintf(out, "Request:  %s %s\n", result.Request.Method, result.Request.URL)
	fmt.Fprintf(out, "Status:   %d %s\n", result.Response.Status, result.Response.StatusText)
	fmt.Fprintf(out, "Duration: %dms\n\n", result.Response.DurationMs)

	fmt.Fprintf(out, "Configured Plugins (execution order):\n")
	tp := tableprinter.New(out)
	tp.SetHeaders("PLUGIN", "PRIORITY")
	for _, p := range result.ConfiguredPlugins {
		tp.AddRow(p.Name, fmt.Sprintf("%d", p.Priority))
	}
	_ = tp.Render()

	if result.Response.UpstreamStatus != "" {
		fmt.Fprintf(out, "\nUpstream Response:  %s\n", result.Response.UpstreamStatus)
	}
	if len(result.Response.ExecutedPlugins) > 0 {
		fmt.Fprintf(out, "Executed Plugins:   %s\n", strings.Join(result.Response.ExecutedPlugins, ", "))
	} else {
		fmt.Fprintf(out, "Executed Plugins:   (enable debug mode in APISIX to see executed plugins)\n")
	}
}

func formatUpstream(upstream interface{}) string {
	if upstream == nil {
		return "(none)"
	}
	b, err := json.Marshal(upstream)
	if err != nil {
		return "(unknown)"
	}
	return string(b)
}

func selectRoute(opts *Options) (string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return "", err
	}

	httpClient, err := opts.Client()
	if err != nil {
		return "", err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	body, err := client.Get("/apisix/admin/routes", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch routes: %s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Route]
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	items := make([]selector.Item, 0, len(resp.List))
	for _, item := range resp.List {
		id := item.Key
		if item.Value.ID != nil {
			id = *item.Value.ID
		}
		if id == "" {
			continue
		}
		label := id
		if item.Value.Name != nil && *item.Value.Name != "" {
			label = fmt.Sprintf("%s (%s)", *item.Value.Name, id)
		}
		items = append(items, selector.Item{ID: id, Label: label})
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no routes found")
	}

	return selector.SelectOne("Select a route", items)
}
