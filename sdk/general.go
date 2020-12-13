package sdk

import (
	"context"
	"fmt"
	"time"
)

// UserInfo contains properties about a user account.
type UserInfo struct {
	result
	CryptoSetup           bool
	Plan                  int
	CryptoSubscription    bool
	UserID                int
	HasPassword           bool
	PublicLinkQuota       uint64
	CryptoLifetime        bool
	PremiumExpires        APITime
	Email                 string
	TrashRevRetentionDays int
	Auth                  string
	EmailVerified         bool
	UsedPublinkBranding   bool
	Currency              string
	AgreedWithPP          bool // pp: privacy policy
	Quota                 uint64
	CryptoExpires         APITime
	Premium               bool
	PremiumLifetime       bool
	Business              bool
	UsedQuota             uint64
	Language              string
	HasPaidRelocation     bool
	Registered            APITime
	RegistrationInfo      RegistrationInfo
	Journey               Journey
	APIServer             APIServer
}

// RegistrationInfo contains registration information about a user account.
type RegistrationInfo struct {
	Provider int
	Device   string
	Country  string
	Ref      int
}

// Journey contains the details of the user registration journey steps.
type Journey struct {
	Steps Steps
}

// Steps contains the various steps of the user registration journey.
type Steps struct {
	VerifyMail     bool
	UploadFile     bool
	AutoUpload     bool
	DownloadApp    bool
	DownloadDrive  bool
	SentInvitation bool
}

// APIServer contains the details of the API servers: Binary and JSON API endpoints.
type APIServer struct {
	BinAPI []string
	API    []string
}

// DiffResult is returned by diff.
type DiffResult struct {
	result
	DiffID  uint64
	Entries []Entry
}

// Entry is a component of DiffResult which is returned by diff.
type Entry struct {
	Event    Event
	Time     APITime
	DiffID   uint64
	Metadata Metadata
}

// UserInfo returns information about the current user.
// As there is no specific login method as credentials can be passed to any method,
// this is an especially good place for logging in with no particular action in mind.
// https://docs.pcloud.com/methods/general/userinfo.html
func (c *Client) UserInfo(ctx context.Context, opts ...ClientOption) (*UserInfo, error) {
	q := toQuery(opts...)

	q.Add("getregistrationinfo", "1")
	q.Add("getapiserver", "1")

	e := &UserInfo{}

	err := parseAPIOutput(e)(c.get(ctx, "userinfo", q))
	if err != nil {
		return nil, err
	}

	return e, nil
}

// Diff lists updates of the user's folders/files.
// Optionally, takes the parameter diffid, which if provided returns only changes since that
// diffid.
// Alternatively you can provide date/time in after parameter and you will only receive events
// generated after that time.
// Another alternative to providing diffid or after is providing last, which will return last
// number of events with highest diffids (that is the last events).
// Especially setting last to 0 is optimized to do nothing more than return the last diffid.
// If the optional parameter block is set and there are no changes since the provided diffid,
// the connection will block until an event arrives. Blocking only works when diffid is provided
// and does not work with either after or last.
// However, sending any additional data on the blocked connection will unblock the request and
// an empty set will be returned. This is useful when you want to monitor for updates when idle
// and use connection for other activities when needed.
// Just keep in mind that if you send any request on a connection that is blocked, you will
// receive two replies - one with empty set of updates and one answering your second request.
// If the optional limit parameter is provided, no more than limit entries will be returned.
// IMPORTANT When a folder/file is created/delete/moved in or out of a folder, you are supposed
// to update modification time of the parent folder to the timestamp of the event.
// IMPORTANT If your state is more than 6 months old, you are advised to re-download all your
// state again, as we reserve the right to compact data that is more than 6 months old.
// Compacting means that if a deletefolder/deletefile event is more than 6 month old, it will
// disappear altogether with all create/modify events. Also, if modifyfile is more than 6 months
// old, it can become createfile and the original createfile will disappear. That is not
// comprehensive list of compacting activities, so you should generally re-download from zero
// rather than trying to cope with compacting.
// https://docs.pcloud.com/methods/general/diff.html
////////////////////////////////
// TODO: add support for shares.
////////////////////////////////
func (c *Client) Diff(ctx context.Context, diffID uint64, after time.Time, last uint64, block bool, limit uint64, opts ...ClientOption) (*DiffResult, error) {
	q := toQuery(opts...)

	if diffID > 0 {
		q.Add("diffid", fmt.Sprintf("%d", diffID))
	}

	if !after.IsZero() {
		q.Add("after", after.Format(ctLayout))
	}

	if last > 0 {
		q.Add("last", fmt.Sprintf("%d", last))
	}

	if block {
		q.Add("block", "1")
	}

	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}

	dr := &DiffResult{}

	err := parseAPIOutput(dr)(c.get(ctx, "diff", q))
	if err != nil {
		return nil, err
	}

	return dr, nil
}
