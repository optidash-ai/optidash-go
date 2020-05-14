package optidash

import (
    "errors"
    "io"
    "net/http"
)

const apiURL = "https://api.optidash.ai/1.0"

// Client is the Optidash HTTP API client
type Client struct {
    Key    string
    Client *http.Client
}

// NewClient returns a new client using the given config.
// The *Config has to be generated using the NewConfig() function.
func NewClient(key string) (*Client, error) {
    // Key must not be empty
    if key == "" {
        return nil, errors.New("Invalid configuration, API key is empty")
    }

    return &Client{
        Key:    key,
        Client: http.DefaultClient,
    }, nil
}

// Upload accepts either a io.Reader (a stream) or a string (a file to read).
// Returns a new API request builder.
func (c *Client) Upload(input interface{}) *Request {
    switch v := input.(type) {
    case io.Reader:
        return &Request{
            client: c,
            http:   c.Client,
            source: readerSource,
            reader: v,
        }
    case string:
        return &Request{
            client:   c,
            http:     c.Client,
            source:   pathSource,
            location: v,
        }
    }

    return nil
}

// Fetch accepts a URL to a resource which the API should download.
// Returns a new API request builder.
func (c *Client) Fetch(url string) *Request {
    return &Request{
        client:   c,
        http:     c.Client,
        source:   fetchSource,
        location: url,
    }
}
