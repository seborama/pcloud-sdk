package sdk

import (
	"context"
	"net/url"
)

type LogoutResult struct {
	result
	AuthDeleted bool
}

// func (r LogoutResult) GetResult() result {
// 	return r.Result
// }

// Logout Gets a token and invalidates it.
// Returns bool auth_deleted if the token invalidation was successful
// (token was correct and it was actually invalidated).
func (c *Client) Logout(ctx context.Context) (*LogoutResult, error) {
	q := url.Values{}
	q.Add("auth", c.auth)

	lr := &LogoutResult{}

	err := parseAPIOutput(lr)(c.request(ctx, "logout", q))
	if err != nil {
		return nil, err
	}

	c.auth = ""

	return lr, nil
}
