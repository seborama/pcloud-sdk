package sdk_test

import (
	"fmt"

	"github.com/google/uuid"
	"seborama/pcloud/sdk"
)

func (testsuite *IntegrationTestSuite) Test_FolderOperations_ByPath() {
	folderPath := testsuite.testFolderPath + "/go_pCloud_" + uuid.New().String()

	_, err := testsuite.pcc.DeleteFolderRecursive(testsuite.ctx, sdk.T1FolderByPath(folderPath))
	testsuite.Require().Error(err)
	testsuite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = testsuite.pcc.DeleteFolder(testsuite.ctx, sdk.T1FolderByPath(folderPath))
	testsuite.Require().Error(err)
	testsuite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = testsuite.pcc.CreateFolder(testsuite.ctx, sdk.T2FolderByPath(folderPath))
	testsuite.Require().NoError(err)

	_, err = testsuite.pcc.CreateFolderIfNotExists(testsuite.ctx, sdk.T2FolderByPath(folderPath))
	testsuite.Require().NoError(err)

	_, err = testsuite.pcc.ListFolder(testsuite.ctx, sdk.T1FolderByPath(folderPath), true, false, false, false)
	testsuite.Require().NoError(err)

	_, err = testsuite.pcc.CreateFolder(testsuite.ctx, sdk.T2FolderByPath(folderPath+" COPY"))
	testsuite.Require().NoError(err)

	_, err = testsuite.pcc.CopyFolder(testsuite.ctx, sdk.T1FolderByPath(folderPath), sdk.ToT1FolderByPath(folderPath+" COPY"), false, false, false)
	testsuite.Require().NoError(err)

	fr, err := testsuite.pcc.DeleteFolderRecursive(testsuite.ctx, sdk.T1FolderByPath(folderPath))
	testsuite.Require().NoError(err)
	testsuite.EqualValues(1, fr.DeletedFolders)
	testsuite.EqualValues(0, fr.DeletedFiles)

	_, err = testsuite.pcc.CreateFolderIfNotExists(testsuite.ctx, sdk.T2FolderByPath(folderPath))
	testsuite.Require().NoError(err)

	lf, err := testsuite.pcc.DeleteFolder(testsuite.ctx, sdk.T1FolderByPath(folderPath))
	testsuite.Require().NoError(err)
	testsuite.Equal(folderPath, lf.Metadata.Path)
}

func (testsuite *IntegrationTestSuite) Test_FolderOperations_ByID() {
	folderName := "go_pCloud_" + uuid.New().String()

	folderPathName := testsuite.testFolderPath + "/" + folderName

	_, err := testsuite.pcc.DeleteFolderRecursive(testsuite.ctx, sdk.T1FolderByPath(folderPathName))
	testsuite.Require().Error(err)
	testsuite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = testsuite.pcc.DeleteFolder(testsuite.ctx, sdk.T1FolderByPath(folderPathName))
	testsuite.Require().Error(err)
	testsuite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	lf, err := testsuite.pcc.CreateFolder(testsuite.ctx, sdk.T2FolderByIDName(testsuite.testFolderID, folderName))
	testsuite.Require().NoError(err)
	folderID := lf.Metadata.FolderID

	_, err = testsuite.pcc.CreateFolderIfNotExists(testsuite.ctx, sdk.T2FolderByIDName(testsuite.testFolderID, folderName))
	testsuite.Require().NoError(err)

	_, err = testsuite.pcc.ListFolder(testsuite.ctx, sdk.T1FolderByID(folderID), true, false, false, false)
	testsuite.Require().NoError(err)

	lf, err = testsuite.pcc.CreateFolder(testsuite.ctx, sdk.T2FolderByIDName(testsuite.testFolderID, folderName+" COPY"))
	testsuite.Require().NoError(err)
	copyFolderID := lf.Metadata.FolderID

	_, err = testsuite.pcc.CopyFolder(testsuite.ctx, sdk.T1FolderByID(folderID), sdk.ToT1FolderByID(copyFolderID), false, false, false)
	testsuite.Require().NoError(err)

	fr, err := testsuite.pcc.DeleteFolderRecursive(testsuite.ctx, sdk.T1FolderByID(folderID))
	testsuite.Require().NoError(err)
	testsuite.EqualValues(1, fr.DeletedFolders)
	testsuite.EqualValues(0, fr.DeletedFiles)

	lf, err = testsuite.pcc.CreateFolderIfNotExists(testsuite.ctx, sdk.T2FolderByIDName(testsuite.testFolderID, folderName))
	testsuite.Require().NoError(err)
	folderID = lf.Metadata.FolderID

	lf, err = testsuite.pcc.DeleteFolder(testsuite.ctx, sdk.T1FolderByID(folderID))
	testsuite.Require().NoError(err)
	testsuite.EqualValues(folderID, lf.Metadata.FolderID)
}
