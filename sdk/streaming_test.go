package sdk_test

import (
	"context"
	"seborama/pcloud/sdk"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_GetFileLink(t *testing.T) {
	folderPath := "/go_pCloud_" + uuid.New().String()
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	_, err := pcc.CreateFolder(context.Background(), folderPath, 0, "")
	require.NoError(t, err)
	defer func() {
		_, err = pcc.DeleteFolderRecursive(context.Background(), folderPath, 0)
		require.NoError(t, err)
	}()

	f, err := pcc.FileOpen(context.Background(), sdk.O_CREAT|sdk.O_EXCL, folderPath+"/"+fileName, 0, 0, "")
	require.NoError(t, err)

	fdt, err := pcc.FileWrite(context.Background(), f.FD, []byte(Lipsum))
	require.NoError(t, err)
	require.EqualValues(t, len(Lipsum), fdt.Bytes)

	err = pcc.FileClose(context.Background(), f.FD)
	require.NoError(t, err)

	fl, err := pcc.GetFileLink(context.Background(), folderPath+"/"+fileName, 0, true, "", 0, false)
	require.NoError(t, err)
	require.Equal(t, 0, fl.Result)
	require.GreaterOrEqual(t, len(fl.Path), 10)
	require.EqualValues(t, '/', fl.Path[0])
	require.True(t, fl.Expires.After(time.Now().Add(time.Hour)))
	require.GreaterOrEqual(t, len(fl.Hosts), 1)
}
