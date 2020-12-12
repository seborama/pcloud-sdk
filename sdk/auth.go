package sdk

import (
	"context"

	"github.com/pkg/errors"
)

// LogoutResult contains the properties returned from an API call to Logout.
type LogoutResult struct {
	result
	AuthDeleted bool
}

// Login performs a user login by credentials supplied via opts.
// Typically this could be username and password.
// This is not an SDK method per-se, rather a wrapper around UserInfo.
// https://docs.pcloud.com/methods/intro/authentication.html
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
func (c *Client) Logout(ctx context.Context, opts ...ClientOption) (*LogoutResult, error) {
	q := toQuery(opts...)

	lr := &LogoutResult{}

	err := parseAPIOutput(lr)(c.request(ctx, "logout", q))
	if err != nil {
		return nil, err
	}

	c.auth = ""

	return lr, nil
}
