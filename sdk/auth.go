package sdk

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// LogoutResult contains the properties returned from an API call to Logout.
type LogoutResult struct {
	result
	AuthDeleted bool
}

// LoginV1 performs a user login by credentials supplied via opts.
// Typically this could be username and password.
// This is not an SDK method per-se, rather a wrapper around UserInfo.
// https://docs.pcloud.com/methods/intro/authentication.html
func (c *Client) LoginV1(ctx context.Context, opts ...ClientOption) error {
	if c.auth != "" {
		return errors.New("Login called while already logged in. Please call Logout first")
	}

	q := toQuery(opts...)

	q.Add("getauth", "1")
	q.Add("logout", "1")

	ui := &UserInfo{}

	err := parseAPIOutput(ui)(c.get(ctx, "userinfo", q))
	if err != nil {
		return err
	}

	c.auth = ui.Auth

	return nil
}

func deviceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Unknown"
	}
	hostname = strings.Title(hostname)
	sysOS := strings.Title(runtime.GOOS)
	sysArch := strings.Title(runtime.GOARCH)

	return fmt.Sprintf("%s, %s, %s, go pCloud SDK", hostname, sysOS, sysArch)
}

func osID() string {
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		return "5"
	case "darwin":
		return "6"
	case "linux":
		return "7"
	default:
		return "0"
	}
}

// Login performs a user login by credentials supplied via opts.
// Login will handle two-factor authentication where applicable.
// Typically this would be by username and password.
// This is not a documented SDK method.
// https://docs.pcloud.com/methods/intro/authentication.html
func (c *Client) Login(ctx context.Context, otpCodeOpt string, opts ...ClientOption) error {
	if c.auth != "" {
		return errors.New("Login called while already logged in. Please call Logout first")
	}

	q := toQuery(opts...)
	fmt.Println("deviceID", deviceID())

	q.Add("getauth", "1")
	q.Add("logout", "1")
	q.Add("cryptokeyssign", "1") // TODO: is this needed?
	// q.Add("getapiserver", "1")
	q.Add("os", osID())
	q.Add("device", deviceID()) // TODO: is this needed?
	q.Add("deviceid", deviceID())

	ui := &UserInfo{}

	err := parseAPIOutput(ui)(c.get(ctx, "login", q))
	if err != nil {
		if ui.Result != ErrTFARequired {
			// TODO: there may be other flows in the login procedure for consideration, such as:
			//       - 2064: expired token
			//       - 2012: invalid code (probably equivalent to bad login: return error)
			//       - 2205 / 2229: something about "auth expired" needs auth reset (?)
			//       - 2237: expired digest??
			return err
		}

		if ui.Token == "" {
			return errors.New("login requires TFA challenge but token is missing from response")
		}
		return c.loginTFA(ctx, ui.Token, otpCodeOpt) // is the Token worth saving in Client and to what purpose?
	}

	c.auth = ui.Auth

	return nil
}

func (c *Client) loginTFA(ctx context.Context, token, otpCode string, opts ...ClientOption) error {
	q := toQuery(opts...)

	q.Add("getauth", "1")
	q.Add("logout", "1")
	// q.Add("getapiserver", "1")
	q.Add("os", osID())
	q.Add("device", deviceID()) // TODO: is this needed?
	q.Add("deviceid", deviceID())
	q.Add("token", token)     // TFA challenge
	q.Add("code", otpCode)    // TFA response
	q.Add("trustdevice", "1") // TODO: make this configurable

	ui := &UserInfo{}

	err := parseAPIOutput(ui)(c.get(ctx, "tfa_login", q))
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

	err := parseAPIOutput(lr)(c.get(ctx, "logout", q))
	if err != nil {
		return nil, err
	}

	c.auth = ""

	return lr, nil
}
