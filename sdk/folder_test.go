package sdk_test

import (
	"fmt"
	"seborama/pcloud/sdk"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) Test_FolderOperations_ByPath() {
	folderPath := suite.testFolderPath + "/go_pCloud_" + uuid.New().String()

	_, err := suite.pcc.DeleteFolderRecursive(suite.ctx, sdk.T1FolderByPath(folderPath))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = suite.pcc.DeleteFolder(suite.ctx, sdk.T1FolderByPath(folderPath))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	lf, err := suite.pcc.CreateFolder(suite.ctx, sdk.T2FolderByPath(folderPath))
	suite.Require().NoError(err)

	_, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, sdk.T2FolderByPath(folderPath))
	suite.Require().NoError(err)

	_, err = suite.pcc.ListFolder(suite.ctx, sdk.T1FolderByPath(folderPath), true, false, false, false)
	suite.Require().NoError(err)

	_, err = suite.pcc.CreateFolder(suite.ctx, sdk.T2FolderByPath(folderPath+" COPY"))
	suite.Require().NoError(err)

	_, err = suite.pcc.CopyFolder(suite.ctx, sdk.T1FolderByPath(folderPath), sdk.ToT1FolderByPath(folderPath+" COPY"), false, false, false)
	suite.Require().NoError(err)

	fr, err := suite.pcc.DeleteFolderRecursive(suite.ctx, sdk.T1FolderByPath(folderPath))
	suite.Require().NoError(err)
	suite.EqualValues(1, fr.DeletedFolders)
	suite.EqualValues(0, fr.DeletedFiles)

	_, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, sdk.T2FolderByPath(folderPath))
	suite.Require().NoError(err)

	lf, err = suite.pcc.DeleteFolder(suite.ctx, sdk.T1FolderByPath(folderPath))
	suite.Require().NoError(err)
	suite.Equal(folderPath, lf.Metadata.Path)
}

func (suite *IntegrationTestSuite) Test_FolderOperations_ByID() {
	folderName := "go_pCloud_" + uuid.New().String()

	folderPathName := suite.testFolderPath + "/" + folderName

	_, err := suite.pcc.DeleteFolderRecursive(suite.ctx, sdk.T1FolderByPath(folderPathName))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = suite.pcc.DeleteFolder(suite.ctx, sdk.T1FolderByPath(folderPathName))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	lf, err := suite.pcc.CreateFolder(suite.ctx, sdk.T2FolderByIDName(suite.testFolderID, folderName))
	suite.Require().NoError(err)
	folderID := lf.Metadata.FolderID

	_, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, sdk.T2FolderByIDName(suite.testFolderID, folderName))
	suite.Require().NoError(err)

	_, err = suite.pcc.ListFolder(suite.ctx, sdk.T1FolderByID(folderID), true, false, false, false)
	suite.Require().NoError(err)

	lf, err = suite.pcc.CreateFolder(suite.ctx, sdk.T2FolderByIDName(suite.testFolderID, folderName+" COPY"))
	suite.Require().NoError(err)
	copyFolderID := lf.Metadata.FolderID

	_, err = suite.pcc.CopyFolder(suite.ctx, sdk.T1FolderByID(folderID), sdk.ToT1FolderByID(copyFolderID), false, false, false)
	suite.Require().NoError(err)

	fr, err := suite.pcc.DeleteFolderRecursive(suite.ctx, sdk.T1FolderByID(folderID))
	suite.Require().NoError(err)
	suite.EqualValues(1, fr.DeletedFolders)
	suite.EqualValues(0, fr.DeletedFiles)

	lf, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, sdk.T2FolderByIDName(suite.testFolderID, folderName))
	suite.Require().NoError(err)
	folderID = lf.Metadata.FolderID

	lf, err = suite.pcc.DeleteFolder(suite.ctx, sdk.T1FolderByID(folderID))
	suite.Require().NoError(err)
	suite.EqualValues(folderID, lf.Metadata.FolderID)
}
