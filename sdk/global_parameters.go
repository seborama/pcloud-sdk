package sdk

// See https://docs.pcloud.com/methods/intro/global_parameters.html

type Options map[string]string

type ClientOptions func(o *Options)

// WithGlobalOptionID if set to anything, you will get it back in the reply (no matter successful or not). This might be useful if you pipeline requests from many places over single connection.
func WithGlobalOptionID(id string) ClientOptions {
	return func(o *Options) {
		(*o)["id"] = id
	}
}

// WithGlobalOptionTimeFormatAsUnixUTCTimestamp if set, all datetime fields will be represented
// as UTC unix timestamps, otherwise the default date format is used.
// The default datetime format is Thu, 21 Mar 2013 18:31:45 +0000 (rfc 2822), exactly 31 bytes
// long.
func WithGlobalOptionTimeFormatAsUnixUTCTimestamp() ClientOptions {
	// TODO: this is currently ineffective as it isn't coordinated with `APITime`
	return func(o *Options) {
		(*o)["timeformat"] = "timestamp"
	}
}

// WithGlobalOptionGetAuth if set, upon successful authentication an auth token will be returned.
// Auth tokens are at most 64 bytes long and can be passed back instead of username/password
// credentials by auth parameter.
// This token is especially good for setting the auth cookie to keep the user logged in.
func WithGlobalOptionGetAuth() ClientOptions {
	return func(o *Options) {
		(*o)["getauth"] = "1"
	}
}

// WithGlobalOptionFilterMeta if set, it is supposed to be a comma (with no whitespace
// after it) separated list of fileds of metadata that you wish to receive from all calls
// returning metadata. This may be used to eliminate fields that you don't use and thus reduce
// the amount of traffic and parsing required for communications.
// If set to empty string/0 restores the default all value.
// You don't need to send this with every request, once per connection suffices.
func WithGlobalOptionFilterMeta() ClientOptions {
	return func(o *Options) {
		(*o)["filtermeta"] = "1"
	}
}
