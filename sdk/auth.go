package sdk

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
)

type LogoutResult struct {
	result
	AuthDeleted bool
}

func (c *Client) Login(ctx context.Context, opts ...ClientOption) error {
	if c.auth != "" {
		return errors.New("Login called while already logged in. Please call Logout first")
	}

	q := toQuery(opts...)
	q.Add("getauth", "1")
	q.Add("logout", "1")

	ui := &UserInfo{}

	err := parseAPIOutput(ui)(c.request(ctx, "userinfo", q))
	if err != nil {
		return err
	}

	c.auth = ui.Auth

	return nil
}

// Logout Gets a token and invalidates it.
// Returns bool auth_deleted if the token invalidation was successful
// (token was correct and it was actually invalidated).
// https://docs.pcloud.com/methods/auth/logout.html
func (c *Client) Logout(ctx context.Context) (*LogoutResult, error) {
	q := url.Values{}

	lr := &LogoutResult{}

	err := parseAPIOutput(lr)(c.request(ctx, "logout", q))
	if err != nil {
		return nil, err
	}

	c.auth = ""

	return lr, nil
}
