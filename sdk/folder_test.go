package sdk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ListFolderByID(t *testing.T) {
	getAuth()

	lf, err := pcc.ListFolder(context.Background(), "", 0, true, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, 0, lf.Result)
}

func Test_ListFolderByPath(t *testing.T) {
	getAuth()

	lf, err := pcc.ListFolder(context.Background(), "/Test", 0, true, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, 0, lf.Result)
}
