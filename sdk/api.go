package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Client contains the data necessary to make API calls to pCloud.
type Client struct {
	httpClient *http.Client
	apiURL     string

	// Auth tokens are at most 64 bytes long and can be passed back instead of username/password
	// credentials by `auth` parameter. This token is especially good for setting the `auth` cookie
	// to keep the user logged in.
	auth string

	lock sync.Mutex
}

// NewClient creates a new initialised pCloud Client.
func NewClient(c *http.Client) *Client {
	return &Client{
		httpClient: c,
		apiURL:     "eapi.pcloud.com", // TODO: have a retry strategy that sets the URL when logon is successful with one of the datacentres (US or EU)
	}
}

// do executes an HTTPS (enforced) request to the pCloud API endpoint.
// it returns the content-type string, the data from the response and an error, if applicable.
func (c *Client) do(ctx context.Context, method string, endpoint string, query url.Values, contentType string, data []byte) (string, []byte, error) {
	if c.auth != "" {
		query.Add("auth", c.auth)
	}

	u := url.URL{
		Scheme:   "https",
		Host:     c.apiURL,
		Path:     endpoint,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(data))
	if err != nil {
		return "", nil, errors.Wrapf(err, "http request: %s", method)
	}

	req.Header.Add("Connection", "Keep-Alive")
	// req.Header.Add("Keep-Alive", "timeout=600, max=1000")
	req.Header.Add("Content-Type", contentType)

	c.lock.Lock()
	defer c.lock.Unlock()

	resp, err := c.httpClient.Do(req)
	if resp != nil {
		defer func() {
			_, err = io.Copy(ioutil.Discard, resp.Body)
			if err != nil {
				fmt.Println("error discarding remainder of response body:", err.Error())
			}
			err = resp.Body.Close()
			if err != nil {
				fmt.Println("error closing the response body:", err.Error())
			}
		}()
	}
	if err != nil {
		return resp.Header.Get("content-type"), nil, errors.Wrap(err, "http Do")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.Header.Get("content-type"), nil, errors.Wrap(err, "body")
	}

	if resp.StatusCode != http.StatusOK {
		return resp.Header.Get("content-type"), nil, errors.New(string(body))
	}

	return resp.Header.Get("content-type"), body, nil
}

// get executes an HTTPS (enforced) GET to the pCloud API endpoint.
func (c *Client) get(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	_, body, err := c.do(ctx, http.MethodGet, endpoint, query, "application/json", nil)
	return body, err
}

// binget executes an HTTPS (enforced) GET to the pCloud API endpoint.
// It differs from get() in that the content-type is expected to be 'application/octet-stream'.
// When the content-type is application/json and the 'X-Error: xxxx' header is present, it
// returns an error instead.
func (c *Client) binget(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	ct, body, err := c.do(ctx, http.MethodGet, endpoint, query, "application/octet-stream", nil)
	if err != nil {
		return nil, err
	}

	if ct == "application/octet-stream" {
		return body, err
	}

	if !strings.HasPrefix(ct, "application/json") {
		return nil, errors.Errorf("internal error: unrecognised content-type: '%s'", ct)
	}

	r := &result{}
	return nil, parseResult(body, nil, r)
}

// put executes an HTTPS (enforced) PUT to the pCloud API endpoint.
func (c *Client) put(ctx context.Context, endpoint string, query url.Values, data []byte) ([]byte, error) {
	_, body, err := c.do(ctx, http.MethodPut, endpoint, query, "application/octet-stream", data)
	return body, err
}

// post executes an HTTPS (enforced) POST with multipart/form-data to the pCloud API endpoint.
func (c *Client) post(ctx context.Context, endpoint string, query url.Values, contentType string, data []byte) ([]byte, error) {
	_, body, err := c.do(ctx, http.MethodPost, endpoint, query, contentType, data)
	if err != nil {
		return nil, err
	}
	return body, err
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
