package sdk

import (
	"fmt"
	"strings"
	"time"
)

// APITime contains a Go time.Time. It is used to provide a custom JSON
// marshaler for the times received from pCloud's APIs.
// https://docs.pcloud.com/structures/datetime.html
type APITime struct {
	time.Time
}

const ctLayout = time.RFC1123Z

// UnmarshalJSON parses the JSON-encoded APITime value and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
// This is an implementation of Go's "json.Unmarshaler" interface.
func (ct *APITime) UnmarshalJSON(v []byte) (err error) {
	s := strings.Trim(string(v), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(ctLayout, s)
	return
}

// MarshalJSON returns the JSON encoding of an APITime.
// This is an implementation of Go's "json.Marshaler" interface.
func (ct *APITime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(ctLayout))), nil
}
