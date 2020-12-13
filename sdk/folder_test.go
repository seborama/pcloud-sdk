package sdk_test

import (
	"context"
	"fmt"
	"seborama/pcloud/sdk"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FolderOperations_ByPath(t *testing.T) {
	folderPath := "/go_pCloud_" + uuid.New().String()

	_, err := pcc.DeleteFolderRecursive(context.Background(), folderPath, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = pcc.DeleteFolder(context.Background(), folderPath, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	lf, err := pcc.CreateFolder(context.Background(), folderPath, 0, "")
	require.NoError(t, err)
	require.Equal(t, 0, lf.Result)

	lf, err = pcc.CreateFolderIfNotExists(context.Background(), folderPath, 0, "")
	require.NoError(t, err)
	require.Equal(t, 0, lf.Result)

	lf, err = pcc.ListFolder(context.Background(), folderPath, 0, true, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, 0, lf.Result)

	fr, err := pcc.DeleteFolderRecursive(context.Background(), folderPath, 0)
	require.NoError(t, err)
	assert.EqualValues(t, 1, fr.DeletedFolders)
	assert.EqualValues(t, 0, fr.DeletedFiles)

	lf, err = pcc.CreateFolderIfNotExists(context.Background(), folderPath, 0, "")
	require.NoError(t, err)
	require.Equal(t, 0, lf.Result)

	lf, err = pcc.DeleteFolder(context.Background(), folderPath, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, lf.Result)
	assert.Equal(t, folderPath, lf.Metadata.Path)
}

func Test_FolderOperations_ByID(t *testing.T) {
	folderName := "go_pCloud_" + uuid.New().String()

	_, err := pcc.DeleteFolderRecursive(context.Background(), "/"+folderName, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = pcc.DeleteFolder(context.Background(), "/"+folderName, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	lf, err := pcc.CreateFolder(context.Background(), "", sdk.RootFolderID, folderName)
	require.NoError(t, err)
	require.Equal(t, 0, lf.Result)
	folderID := lf.Metadata.FolderID

	lf, err = pcc.CreateFolderIfNotExists(context.Background(), "", sdk.RootFolderID, folderName)
	require.NoError(t, err)
	require.Equal(t, 0, lf.Result)

	lf, err = pcc.ListFolder(context.Background(), "", folderID, true, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, 0, lf.Result)

	fr, err := pcc.DeleteFolderRecursive(context.Background(), "", folderID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, fr.DeletedFolders)
	assert.EqualValues(t, 0, fr.DeletedFiles)

	lf, err = pcc.CreateFolderIfNotExists(context.Background(), "", sdk.RootFolderID, folderName)
	require.NoError(t, err)
	require.Equal(t, 0, lf.Result)
	folderID = lf.Metadata.FolderID

	lf, err = pcc.DeleteFolder(context.Background(), "", folderID)
	require.NoError(t, err)
	assert.Equal(t, 0, lf.Result)
	assert.EqualValues(t, folderID, lf.Metadata.FolderID)
}
