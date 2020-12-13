package sdk_test

import (
	"context"
	"seborama/pcloud/sdk"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_FileOps_ByPath(t *testing.T) {
	folderPath := "/go_pCloud_" + uuid.New().String()
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	_, err := pcc.CreateFolder(context.Background(), folderPath, 0, "")
	require.NoError(t, err)

	f, err := pcc.FileOpen(context.Background(), sdk.O_CREAT|sdk.O_EXCL, folderPath+"/"+fileName, 0, 0, "")
	require.NoError(t, err)

	fdt, err := pcc.FileWrite(context.Background(), f.FD, []byte(Lipsum))
	require.NoError(t, err)
	require.EqualValues(t, len(Lipsum), fdt.Bytes)

	err = pcc.FileClose(context.Background(), f.FD)
	require.NoError(t, err)

	// TODO: add FileDelete (when available)

	_, err = pcc.DeleteFolderRecursive(context.Background(), folderPath, 0)
	require.NoError(t, err)
}

func Test_FileOpen_FileClose_ByFileID(t *testing.T) {
	t.Skip() // not yet written
	f, err := pcc.FileOpen(context.Background(), 0, "", 0, 0, "Test File 1.pdf")
	require.NoError(t, err)
	require.GreaterOrEqual(t, f.FD, uint64(1))
	require.GreaterOrEqual(t, f.FileID, uint64(1))

	err = pcc.FileClose(context.Background(), f.FD)
	require.NoError(t, err)
}
