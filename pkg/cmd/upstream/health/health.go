package health

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
	Output     string
	ControlURL string
}

type HealthCheckResponse struct {
	Nodes []HealthCheckNode `json:"nodes"`
	Type  string            `json:"type"`
	Name  string            `json:"name"`
}

type HealthCheckNode struct {
	IP      string             `json:"ip"`
	Port    int                `json:"port"`
	Status  string             `json:"status"`
	Counter HealthCheckCounter `json:"counter"`
}

type HealthCheckCounter struct {
	Success        int `json:"success"`
	HTTPFailure    int `json:"http_failure"`
	TCPFailure     int `json:"tcp_failure"`
	TimeoutFailure int `json:"timeout_failure"`
}

func NewCmdHealth(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
		Client: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "health [id]",
		Short: "Show health check status of upstream nodes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ID = args[0]
			}
			return healthRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: table, json, yaml")
	cmd.Flags().StringVar(&opts.ControlURL, "control-url", "", "APISIX Control API URL")

	return cmd
}

func healthRun(opts *Options) error {
	if opts.ID == "" {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("id argument is required (or run interactively in a terminal)")
		}
		id, err := selectUpstream(opts)
		if err != nil {
			return err
		}
		opts.ID = id
	}

	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	controlURL := opts.ControlURL
	if controlURL == "" {
		controlURL, err = deriveControlURL(cfg.BaseURL())
		if err != nil {
			return fmt.Errorf("failed to derive control API URL: %w", err)
		}
	}

	var httpClient *http.Client
	if opts.ControlClient != nil {
		httpClient, err = opts.ControlClient()
		if err != nil {
			return err
		}
	} else {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	healthURL, err := buildHealthURL(controlURL, opts.ID)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query Control API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("No health check data available for upstream %s. The upstream must have health checks configured and have served at least one request.", opts.ID)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			return fmt.Errorf("Control API request failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("Control API request failed with status %d: %s", resp.StatusCode, msg)
	}

	var healthResp HealthCheckResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
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
		tp := tableprinter.New(opts.IO.Out)
		tp.SetHeaders("NODE", "STATUS", "SUCCESS", "HTTP_FAIL", "TCP_FAIL", "TIMEOUT")
		for _, node := range healthResp.Nodes {
			tp.AddRow(
				fmt.Sprintf("%s:%d", node.IP, node.Port),
				node.Status,
				fmt.Sprintf("%d", node.Counter.Success),
				fmt.Sprintf("%d", node.Counter.HTTPFailure),
				fmt.Sprintf("%d", node.Counter.TCPFailure),
				fmt.Sprintf("%d", node.Counter.TimeoutFailure),
			)
		}
		if err := tp.Render(); err != nil {
			return err
		}
		if healthResp.Type != "" {
			fmt.Fprintf(opts.IO.Out, "\nType: %s\n", healthResp.Type)
		}
		return nil
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(healthResp)
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

func buildHealthURL(controlBaseURL, id string) (string, error) {
	u, err := url.Parse(controlBaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid control URL: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid control URL: missing host")
	}
	u.Path = "/v1/healthcheck/upstreams/" + id
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func selectUpstream(opts *Options) (string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return "", err
	}

	if opts.Client == nil {
		return "", fmt.Errorf("client is not configured")
	}

	httpClient, err := opts.Client()
	if err != nil {
		return "", err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())
	body, err := client.Get("/apisix/admin/upstreams", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch upstreams: %s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Upstream]
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
		return "", fmt.Errorf("no upstreams found")
	}

	return selector.SelectOne("Select an upstream", items)
}
