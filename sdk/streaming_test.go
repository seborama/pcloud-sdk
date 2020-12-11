package sdk_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetFileLink(t *testing.T) {
	fl, err := pcc.GetFileLink(context.Background(), "/Test/My Folder/File 1.pdf", 0, true, "", 0, false)
	require.NoError(t, err)
	assert.Equal(t, 0, fl.Result)
	assert.GreaterOrEqual(t, len(fl.Path), 10)
	assert.EqualValues(t, '/', fl.Path[0])
	assert.True(t, fl.Expires.After(time.Now().Add(time.Hour)))
	assert.GreaterOrEqual(t, len(fl.Hosts), 1)
}
