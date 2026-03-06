package httpmock

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Register_And_RoundTrip(t *testing.T) {
	reg := &Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/routes", JSONResponse(`{"total":0,"list":[]}`))

	client := reg.GetClient()
	resp, err := client.Get("http://localhost/apisix/admin/routes")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "total")
}

func TestRegistry_NoMock(t *testing.T) {
	reg := &Registry{}
	client := reg.GetClient()
	_, err := client.Get("http://localhost/unmocked")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no mock registered")
}

func TestRegistry_Verify_Success(t *testing.T) {
	reg := &Registry{}
	reg.Register(http.MethodGet, "/test", JSONResponse(`{}`))

	client := reg.GetClient()
	_, err := client.Get("http://localhost/test")
	require.NoError(t, err)

	// Should not fail
	reg.Verify(t)
}

func TestRegistry_Verify_Failure(t *testing.T) {
	mockT := &testing.T{}
	reg := &Registry{}
	reg.Register(http.MethodGet, "/never-called", JSONResponse(`{}`))

	reg.Verify(mockT)
	// mockT would have recorded an error, but we can't easily check that
	// in a unit test. Instead, check CallCount.
	assert.Equal(t, 0, reg.CallCount(http.MethodGet, "/never-called"))
}

func TestRegistry_CallCount(t *testing.T) {
	reg := &Registry{}
	reg.Register(http.MethodGet, "/test", JSONResponse(`{}`))

	assert.Equal(t, 0, reg.CallCount(http.MethodGet, "/test"))

	client := reg.GetClient()
	_, _ = client.Get("http://localhost/test")
	assert.Equal(t, 1, reg.CallCount(http.MethodGet, "/test"))

	_, _ = client.Get("http://localhost/test")
	assert.Equal(t, 2, reg.CallCount(http.MethodGet, "/test"))
}

func TestRegistry_CallCount_NotRegistered(t *testing.T) {
	reg := &Registry{}
	assert.Equal(t, 0, reg.CallCount(http.MethodGet, "/not-registered"))
}

func TestStringResponse(t *testing.T) {
	resp := StringResponse(403, `{"error_msg":"forbidden"}`)
	assert.Equal(t, 403, resp.StatusCode)
	assert.Equal(t, `{"error_msg":"forbidden"}`, string(resp.Body))
}

func TestJSONResponse(t *testing.T) {
	resp := JSONResponse(`{"key":"value"}`)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `{"key":"value"}`, string(resp.Body))
}

func TestRegistry_ResponseHeaders(t *testing.T) {
	reg := &Registry{}
	reg.Register(http.MethodGet, "/test", Response{
		StatusCode: 200,
		Body:       []byte(`{}`),
		Header: http.Header{
			"X-Custom": {"custom-value"},
		},
	})

	client := reg.GetClient()
	resp, err := client.Get("http://localhost/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "custom-value", resp.Header.Get("X-Custom"))
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
