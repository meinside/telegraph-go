package telegraph

// http://telegra.ph/api

// constants
const (
	apiBaseURL = "https://api.telegra.ph"
)

// Verbose flag for logging
var Verbose bool // default: false

// Client struct
type Client struct {
	AccessToken string
}

// Create creates a new Telegraph client.
func Create(shortName, authorName, authorURL string) (client *Client, err error) {
	var account Account
	if account, err = client.CreateAccount(shortName, authorName, authorURL); err == nil {
		client = &Client{AccessToken: account.AccessToken}
	}

	return client, err
}

// Load a Telegraph client with an existing access token.
func Load(accessToken string) (client *Client, err error) {
	client = &Client{AccessToken: accessToken}

	_, err = client.GetAccountInfo(nil)

	return client, err
}
