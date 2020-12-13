package sdk

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Client contains the data necessary to make API calls to pCloud.
type Client struct {
	apiURL string

	// Auth tokens are at most 64 bytes long and can be passed back instead of username/password
	// credentials by `auth` parameter. This token is especially good for setting the `auth` cookie
	// to keep the user logged in.
	auth string
}

// NewClient creates a new initialised pCloud Client.
func NewClient() *Client {
	return &Client{
		apiURL: "eapi.pcloud.com", // TODO: have a retry strategy that sets the URL when logon is successful with one of the datacentres (US or EU)
	}
}

// get executes an HTTPS (enforced) request to the pCloud API endpoint.
func (c *Client) do(ctx context.Context, method string, endpoint string, query url.Values, data []byte) ([]byte, error) {
	if c.auth != "" {
		query.Add("auth", c.auth)
	}

	u := url.URL{
		Scheme:   "https",
		Host:     c.apiURL,
		Path:     endpoint,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "http request: %s", method)
	}

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "http Do")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	return body, nil
}

// get executes an HTTPS (enforced) GET to the pCloud API endpoint.
func (c *Client) get(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	return c.do(ctx, http.MethodGet, endpoint, query, nil)
}

// put executes an HTTPS (enforced) PUT to the pCloud API endpoint.
func (c *Client) put(ctx context.Context, endpoint string, query url.Values, data []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPut, endpoint, query, data)
}

type result struct {
	Result int    `json:"result"`
	Error  string `json:"error"`
}

// Result_ returns the Result property.
func (r result) Result_() int { return r.Result }

// Error_ returns the Error property.
func (r result) Error_() string { return r.Error }

type resulter interface {
	Result_() int
	Error_() string
}

type resultGetter interface {
	GetResult() result
}

// parseAPIOutput is a curry for parseResult.
func parseAPIOutput(r resulter) func(body []byte, err error) error {
	return func(body []byte, err error) error {
		return parseResult(body, err, r)
	}
}

// parseResult parses the body of the response from the pCloud API.
func parseResult(body []byte, err error, r resulter) error {
	if err != nil {
		return errors.WithStack(err)
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return errors.Wrap(err, "unmarshal")
	}
	if r.Result_() != 0 {
		return errors.Errorf("error %d: %s", r.Result_(), r.Error_())
	}
	return nil
}

// toQuery create a blank url.Value object for use as a query with an HTTPS request.
// It applies the options specified by opts to it and returns it.
func toQuery(opts ...ClientOption) url.Values {
	q := url.Values{}

	for _, opt := range opts {
		opt(&q)
	}

	return q
}
