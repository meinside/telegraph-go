package telegraph

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidParameter is wrapped by errors for parameters rejected before any request is sent.
var ErrInvalidParameter = errors.New("telegraph: invalid parameter")

func invalidParamf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidParameter, fmt.Sprintf(format, args...))
}

// APIError is returned when the API answers with "ok": false.
type APIError struct {
	Method string // API method, e.g. "createPage"
	Code   string // error from the API, e.g. "PAGE_NOT_FOUND"
}

func (e *APIError) Error() string {
	return fmt.Sprintf("telegraph: %s: %s", e.Method, e.Code)
}

// FloodWait reports how long to wait before retrying when the API answered FLOOD_WAIT_<seconds>.
func (e *APIError) FloodWait() (time.Duration, bool) {
	rest, ok := strings.CutPrefix(e.Code, "FLOOD_WAIT_")
	if !ok {
		return 0, false
	}
	secs, err := strconv.Atoi(rest)
	if err != nil || secs < 0 {
		return 0, false
	}
	return time.Duration(secs) * time.Second, true
}

// HTTPError is returned when the API answers with a non-2xx status and no API error in the body.
type HTTPError struct {
	Method     string
	StatusCode int
	Body       string // truncated to maxErrorBodyBytes
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("telegraph: %s: HTTP %d: %s", e.Method, e.StatusCode, e.Body)
}
