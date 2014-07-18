package digitalocean

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const DIGITALOCEAN_API_URL = "https://api.digitalocean.com/v2"

// Client provides a client to the DigitalOcean API
type Client struct {
	// Access Token
	Token string

	// URL to the DO API to use
	URL string

	// HttpClient is the client to use. Default will be
	// used if not provided.
	Http *http.Client
}

// DoError is the error format that they return
// to us if there is a problem
type DoError struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

// NewClient returns a new digitalocean client,
// requires an authorization token. You can generate
// an OAuth token by visiting the Apps & API section
// of the DigitalOcean control panel for your account.
func NewClient(token string) (*Client, error) {
	client := &Client{
		Token: token,
		URL:   DIGITALOCEAN_API_URL,
	}
	return client, nil
}

// Creates a new request with the params
func (c *Client) NewRequest(params map[string]string, method string, endpoint string) (*http.Request, error) {
	p := url.Values{}
	u, err := url.Parse(c.URL)

	if err != nil {
		return nil, fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Build up our request parameters
	for k, v := range params {
		p.Add(k, v)
	}

	// Add the params to our URL
	u.RawQuery = p.Encode()

	// Build the request
	req, err := http.NewRequest(method, u.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("Error creating request: %s", err)
	}

	// Add the authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	return req, nil

}

// parseErr is used to take an error json resp
// and return a single string for use in error messages
func parseErr(resp *http.Response) error {
	errBody := new(DoError)

	err := decodeBody(resp, errBody)

	// if there was an error decoding the body, just return that
	if err != nil {
		return fmt.Errorf("Error parsing error body for non-200 request: %s", err)
	}

	return fmt.Errorf("API Error: %s: %s", errBody.Id, errBody.Message)
}

// decodeBody is used to JSON decode a body
func decodeBody(resp *http.Response, out interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &out); err != nil {
		return err
	}

	return nil
}

// checkResp wraps http.Client.Do() and verifies that the
// request was successful. A non-200 request returns an error
// formatted to included any validation problems or otherwise
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	// If the err is already there, there was an error higher
	// up the chain, so just return that
	if err != nil {
		return resp, err
	}

	// Verify that the request was sucessful
	// 200 is the standard request code returned by the DO API,
	// but 204 is used on successful DELETE requests
	if resp.StatusCode != 200 || resp.StatusCode != 204 {
		// Parse the err and retun it
		return resp, parseErr(resp)
	}

	// The request was succesful, so return a nil error
	return resp, nil
}
