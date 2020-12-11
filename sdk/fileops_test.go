package sdk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FileOpen_FileClose_ByPath(t *testing.T) {
	f, err := pcc.FileOpen(context.Background(), 0, "/Test/My Folder/File 1.pdf", 0, 0, "File 1.pdf")
	require.NoError(t, err)
	assert.Equal(t, 0, f.Result)
	assert.GreaterOrEqual(t, f.FD, uint64(1))
	assert.GreaterOrEqual(t, f.FileID, uint64(1))

	err = pcc.FileClose(context.Background(), f.FD)
	require.NoError(t, err)
	assert.Equal(t, 0, f.Result)
}

func Test_FileOpen_FileClose_ByFileID(t *testing.T) {
	t.Skip() // not yet written
	f, err := pcc.FileOpen(context.Background(), 0, "", 0, 0, "Test File 1.pdf")
	require.NoError(t, err)
	assert.Equal(t, 0, f.Result)
	assert.GreaterOrEqual(t, f.FD, uint64(1))
	assert.GreaterOrEqual(t, f.FileID, uint64(1))

	err = pcc.FileClose(context.Background(), f.FD)
	require.NoError(t, err)
	assert.Equal(t, 0, f.Result)
}
