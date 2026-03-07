package api

import (
	"encoding/json"
	"fmt"
)

// ListResponse is the generic response for list endpoints.
// APISIX wraps each item in a struct with a "value" field.
type ListResponse[T any] struct {
	Total int           `json:"total"`
	List  []ListItem[T] `json:"list"`
}

// UnmarshalJSON handles APISIX returning {"list": {}} (empty object) for
// empty results instead of the standard {"list": []} (empty array).
func (r *ListResponse[T]) UnmarshalJSON(data []byte) error {
	var raw struct {
		Total int             `json:"total"`
		List  json.RawMessage `json:"list"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Total = raw.Total

	if len(raw.List) == 0 || string(raw.List) == "null" || string(raw.List) == "{}" {
		r.List = nil
		return nil
	}

	return json.Unmarshal(raw.List, &r.List)
}

// ListItem wraps a single resource in a list response.
type ListItem[T any] struct {
	Key           string `json:"key"`
	Value         T      `json:"value"`
	CreatedIndex  int    `json:"createdIndex"`
	ModifiedIndex int    `json:"modifiedIndex"`
}

// SingleResponse is the generic response for get/create/update endpoints.
type SingleResponse[T any] struct {
	Key           string `json:"key"`
	Value         T      `json:"value"`
	CreatedIndex  int    `json:"createdIndex"`
	ModifiedIndex int    `json:"modifiedIndex"`
}

// APIError represents an error response from the APISIX Admin API.
type APIError struct {
	StatusCode int    `json:"-"`
	ErrorMsg   string `json:"error_msg"`
}

func (e *APIError) Error() string {
	if e.ErrorMsg != "" {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.ErrorMsg)
	}
	return fmt.Sprintf("API error: status %d", e.StatusCode)
}

// DeleteResponse is the response for delete endpoints.
type DeleteResponse struct {
	Key     string `json:"key"`
	Deleted string `json:"deleted"`
}
