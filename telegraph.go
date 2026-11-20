// Package telegraph is a client for the Telegraph API.
//
// https://telegra.ph/api
package telegraph

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DefaultBaseURL is the Telegraph API endpoint.
const DefaultBaseURL = "https://api.telegra.ph"

// defaultTimeout bounds a whole request (connect, send, read) when the caller
// does not supply its own *http.Client. Use a context for per-call deadlines.
const defaultTimeout = 30 * time.Second

// Client calls the Telegraph API. It is safe for concurrent use.
// Get one from NewClient, Create or Load; the zero value is not usable.
type Client struct {
	mu          sync.RWMutex
	accessToken string

	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient makes the client send requests through hc.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithBaseURL replaces DefaultBaseURL, e.g. with an httptest server.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// WithLogger makes the client log each request at debug level.
// The access token is redacted from logged parameters.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		if l != nil {
			c.logger = l
		}
	}
}

// NewClient returns a client for accessToken without contacting the API.
// accessToken may be empty for methods that do not need one (CreateAccount, GetPage, GetViews).
func NewClient(accessToken string, opts ...Option) *Client {
	c := &Client{
		accessToken: accessToken,
		httpClient:  &http.Client{Timeout: defaultTimeout},
		baseURL:     DefaultBaseURL,
		logger:      slog.New(slog.DiscardHandler),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Create creates a new Telegraph account and returns a client for it.
func Create(ctx context.Context, params CreateAccountParams, opts ...Option) (*Client, error) {
	c := NewClient("", opts...)
	account, err := c.CreateAccount(ctx, params)
	if err != nil {
		return nil, err
	}
	c.setAccessToken(account.AccessToken)
	return c, nil
}

// Load returns a client for an existing access token after checking it with GetAccountInfo.
func Load(ctx context.Context, accessToken string, opts ...Option) (*Client, error) {
	c := NewClient(accessToken, opts...)
	if _, err := c.GetAccountInfo(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// AccessToken returns the client's current access token.
// It changes after a successful RevokeAccessToken.
func (c *Client) AccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken
}

func (c *Client) setAccessToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = token
}
