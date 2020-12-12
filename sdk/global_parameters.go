package sdk

import (
	"fmt"
	"net/url"
	"time"
)

// ClientOption is a Go functional parameter signature.
// This is used by most SDK methods to pass global parameters such as username,
// getauth,id, authexpire, etc.
type ClientOption func(q *url.Values)

// WithGlobalOptionID if set to anything, you will get it back in the reply (no matter
// successful or not). This might be useful if you pipeline requests from many places over
// single connection.
// https://docs.pcloud.com/methods/intro/global_parameters.html
func WithGlobalOptionID(id string) ClientOption {
	return func(q *url.Values) {
		q.Add("id", id)
	}
}

// WithGlobalOptionTimeFormatAsUnixUTCTimestamp if set, all datetime fields will be represented
// as UTC unix timestamps, otherwise the default date format is used.
// The default datetime format is Thu, 21 Mar 2013 18:31:45 +0000 (rfc 2822), exactly 31 bytes
// long.
// https://docs.pcloud.com/methods/intro/global_parameters.html
func WithGlobalOptionTimeFormatAsUnixUTCTimestamp() ClientOption {
	// TODO: this is currently ineffective as it isn't coordinated with `APITime`
	return func(q *url.Values) {
		q.Add("timeformat", "timestamp")
	}
}

// WithGlobalOptionGetAuth if set, upon successful authentication an auth token will be returned.
// Auth tokens are at most 64 bytes long and can be passed back instead of username/password
// credentials by auth parameter.
// This token is especially good for setting the auth cookie to keep the user logged in.
// https://docs.pcloud.com/methods/intro/global_parameters.html
func WithGlobalOptionGetAuth() ClientOption {
	return func(q *url.Values) {
		q.Add("getauth", "1")
	}
}

// WithGlobalOptionUsername sets the username in plain text.
// Should only be used over SSL connections.
// https://docs.pcloud.com/methods/intro/global_parameters.html
func WithGlobalOptionUsername(username string) ClientOption {
	return func(q *url.Values) {
		q.Add("username", username)
	}
}

// WithGlobalOptionPassword sets the password in plain text.
// Should only be used over SSL connections.
// https://docs.pcloud.com/methods/intro/global_parameters.html
func WithGlobalOptionPassword(password string) ClientOption {
	return func(q *url.Values) {
		q.Add("password", password)
	}
}

// WithGlobalOptionAuthExpire defines the expire value of authentication token, when it is
// requested. This field is in seconds and the expire will the moment after these seconds
// since the current moment.
// Defaults to 31536000 and its maximum is 63072000.
// https://docs.pcloud.com/methods/intro/global_parameters.html
func WithGlobalOptionAuthExpire(authExpire time.Duration) ClientOption {
	return func(q *url.Values) {
		e := int64(authExpire.Seconds())
		if e < 0 {
			e = 31536000
		}
		if e > 63072000 {
			e = 63072000
		}
		q.Add("authexpire", fmt.Sprintf("%d", e))
	}
}

// WithGlobalOptionAuthInactiveExpire defines the expire_inactive value of authentication token,
// when it is requested. This field is in seconds and the expire_incative will the moment
// after these seconds since the current moment. Defaults to 2678400 and its maximum is 5356800.
// https://docs.pcloud.com/methods/intro/global_parameters.html
func WithGlobalOptionAuthInactiveExpire(authInactiveExpire time.Duration) ClientOption {
	return func(q *url.Values) {
		e := int64(authInactiveExpire.Seconds())
		if e < 0 {
			e = 2678400
		}
		if e > 5356800 {
			e = 5356800
		}
		q.Add("authinactiveexpire", fmt.Sprintf("%d", e))
	}
}
