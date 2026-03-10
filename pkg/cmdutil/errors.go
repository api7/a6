package cmdutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/api7/a6/pkg/api"
)

// SilentError is an error that has already been printed to the user.
// Commands returning this error should not print it again.
type SilentError struct {
	Err error
}

func (e *SilentError) Error() string {
	return e.Err.Error()
}

func (e *SilentError) Unwrap() error {
	return e.Err
}

// IsSilent returns true if the error is a SilentError.
func IsSilent(err error) bool {
	var se *SilentError
	return errors.As(err, &se)
}

// FlagError is an error resulting from invalid flag usage. cobra
// will print usage help when this error type is returned.
type FlagError struct {
	Err error
}

func (e *FlagError) Error() string {
	return e.Err.Error()
}

func (e *FlagError) Unwrap() error {
	return e.Err
}

// FormatAPIError formats an API error for user display.
func FormatAPIError(err error) string {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401:
			return "authentication failed: check your API key"
		case 403:
			return fmt.Sprintf("permission denied: %s", apiErr.ErrorMsg)
		case 404:
			return "resource not found"
		case 409:
			return fmt.Sprintf("conflict: %s", apiErr.ErrorMsg)
		default:
			return apiErr.Error()
		}
	}
	return err.Error()
}

// IsNotFound returns true if the error is a 404 API error.
func IsNotFound(err error) bool {
	var apiErr *api.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

// IsOptionalResourceError returns true if the error indicates
// a resource type is unavailable (e.g., stream mode disabled returns 400,
// or the resource endpoint returns 404). Used to gracefully skip
// optional resources like stream_routes, protos, and secrets during
// config dump/sync.
func IsOptionalResourceError(err error) bool {
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == 400 || apiErr.StatusCode == 404
}

// NormalizeLabel converts "key=value" to "key:value" for the APISIX Admin API.
func NormalizeLabel(label string) string {
	if label == "" {
		return ""
	}
	parts := strings.SplitN(label, "=", 2)
	if len(parts) == 2 {
		return parts[0] + ":" + parts[1]
	}
	return label
}
