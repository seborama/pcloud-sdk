package sdk_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_UserInfo(t *testing.T) {
	ui, err := pcc.UserInfo(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, ui.APIServer)
	require.NotEmpty(t, ui.Email)
}

func Test_Diff(t *testing.T) {
	dr, err := pcc.Diff(context.Background(), 0, time.Now().Add(-time.Hour), 0, false, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, dr.DiffID, uint64(1))
	require.GreaterOrEqual(t, dr.Entries[0].DiffID, uint64(1))
	require.NotEmpty(t, dr.Entries[0].Metadata.Name)
}
