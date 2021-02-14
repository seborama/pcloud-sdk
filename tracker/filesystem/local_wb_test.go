package filesystem

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashData(t *testing.T) {
	h, err := hashData(bytes.NewReader([]byte("This is File000")))
	require.NoError(t, err)
	assert.Equal(t, "01ce643e7c1ca98f6fb21e61b5d03f547813edae", h)
}
