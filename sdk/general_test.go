package sdk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_UserInfo(t *testing.T) {
	ui, err := pcc.UserInfo(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, ui.APIServer)
	require.NotEmpty(t, ui.Email)
}
