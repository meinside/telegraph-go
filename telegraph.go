package telegraph

// http://telegra.ph/api

// constants
const (
	APIBaseURL = "https://api.telegra.ph"
)

// Verbose flag for logging
var Verbose bool // default: false

// Client struct
type Client struct {
	AccessToken string
}

// Create creates a new Telegraph client.
func Create(shortName, authorName, authorURL string) (*Client, error) {
	client := Client{}

	created, err := client.CreateAccount(shortName, authorName, authorURL)

	if err == nil {
		return &Client{AccessToken: created.AccessToken}, nil
	}

	return nil, err
}

// Load a Telegraph client with access token.
func Load(accessToken string) (*Client, error) {
	client := Client{AccessToken: accessToken}

	_, err := client.GetAccountInfo(nil)

	if err == nil {
		return &client, nil
	}

	return nil, err
}
