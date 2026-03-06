package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/apisix/admin/routes", r.URL.Path)
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total":0,"list":[]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	body, err := client.Get("/apisix/admin/routes", map[string]string{"foo": "bar"})
	require.NoError(t, err)
	assert.Contains(t, string(body), "total")
}

func TestClient_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"key":"/apisix/routes/1"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	body, err := client.Post("/apisix/admin/routes", map[string]string{"uri": "/test"})
	require.NoError(t, err)
	assert.Contains(t, string(body), "key")
}

func TestClient_Put(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"/apisix/routes/1"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	body, err := client.Put("/apisix/admin/routes/1", map[string]string{"uri": "/test"})
	require.NoError(t, err)
	assert.Contains(t, string(body), "key")
}

func TestClient_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"/apisix/routes/1"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	body, err := client.Patch("/apisix/admin/routes/1", map[string]string{"name": "updated"})
	require.NoError(t, err)
	assert.Contains(t, string(body), "key")
}

func TestClient_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("force"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"/apisix/routes/1","deleted":"1"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	body, err := client.Delete("/apisix/admin/routes/1", map[string]string{"force": "true"})
	require.NoError(t, err)
	assert.Contains(t, string(body), "deleted")
}

func TestClient_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error_msg":"forbidden"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	_, err := client.Get("/apisix/admin/routes", nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	assert.Equal(t, "forbidden", apiErr.ErrorMsg)
}

func TestClient_APIErrorNoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	_, err := client.Get("/apisix/admin/routes", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal server error")
}

func TestNewAuthenticatedClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-123", r.Header.Get("X-API-KEY"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	httpClient := NewAuthenticatedClient("test-key-123")
	client := NewClient(httpClient, srv.URL)
	_, err := client.Get("/test", nil)
	require.NoError(t, err)
}

func TestClient_GetNilQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	_, err := client.Get("/test", nil)
	require.NoError(t, err)
}

func TestClient_PostNilBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), srv.URL)
	_, err := client.Post("/test", nil)
	require.NoError(t, err)
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      APIError
		expected string
	}{
		{
			name:     "with message",
			err:      APIError{StatusCode: 404, ErrorMsg: "not found"},
			expected: "API error (status 404): not found",
		},
		{
			name:     "without message",
			err:      APIError{StatusCode: 500},
			expected: "API error: status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestListResponse_Unmarshal(t *testing.T) {
	data := `{
		"total": 1,
		"list": [
			{
				"key": "/apisix/routes/1",
				"value": {"name": "test"},
				"createdIndex": 1,
				"modifiedIndex": 2
			}
		]
	}`

	type TestResource struct {
		Name string `json:"name"`
	}

	var resp ListResponse[TestResource]
	err := json.Unmarshal([]byte(data), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "test", resp.List[0].Value.Name)
	assert.Equal(t, "/apisix/routes/1", resp.List[0].Key)
}
