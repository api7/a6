package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a thin wrapper around net/http for the APISIX Admin API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new API client.
func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// NewAuthenticatedClient creates an http.Client with API key authentication.
func NewAuthenticatedClient(apiKey string) *http.Client {
	return &http.Client{
		Transport: &apiKeyTransport{
			apiKey: apiKey,
			base:   http.DefaultTransport,
		},
	}
}

// apiKeyTransport injects the X-API-KEY header into every request.
type apiKeyTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid mutating the original.
	r2 := req.Clone(req.Context())
	r2.Header.Set("X-API-KEY", t.apiKey)
	return t.base.RoundTrip(r2)
}

// Get performs a GET request and returns the response body.
func (c *Client) Get(path string, query map[string]string) ([]byte, error) {
	return c.do(http.MethodGet, path, query, nil)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	return c.do(http.MethodPost, path, nil, body)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	return c.do(http.MethodPut, path, nil, body)
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(path string, body interface{}) ([]byte, error) {
	return c.do(http.MethodPatch, path, nil, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string, query map[string]string) ([]byte, error) {
	return c.do(http.MethodDelete, path, query, nil)
}

func (c *Client) do(method, path string, query map[string]string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if query != nil {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		apiErr.StatusCode = resp.StatusCode
		_ = json.Unmarshal(respBody, &apiErr)
		if apiErr.ErrorMsg == "" {
			apiErr.ErrorMsg = string(respBody)
		}
		return nil, &apiErr
	}

	return respBody, nil
}
