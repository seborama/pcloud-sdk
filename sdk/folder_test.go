package sdk_test

import (
	"fmt"
	"seborama/pcloud/sdk"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) Test_FolderOperations_ByPath() {
	folderPath := suite.testFolderPath + "/go_pCloud_" + uuid.New().String()

	_, err := suite.pcc.DeleteFolderRecursive(suite.ctx, folderPath, 0)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = suite.pcc.DeleteFolder(suite.ctx, folderPath, 0)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	lf, err := suite.pcc.CreateFolder(suite.ctx, folderPath, 0, "")
	suite.Require().NoError(err)

	lf, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, folderPath, 0, "")
	suite.Require().NoError(err)

	lf, err = suite.pcc.ListFolder(suite.ctx, folderPath, 0, true, false, false, false)
	suite.Require().NoError(err)

	fr, err := suite.pcc.DeleteFolderRecursive(suite.ctx, folderPath, 0)
	suite.Require().NoError(err)
	suite.EqualValues(1, fr.DeletedFolders)
	suite.EqualValues(0, fr.DeletedFiles)

	lf, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, folderPath, 0, "")
	suite.Require().NoError(err)

	lf, err = suite.pcc.DeleteFolder(suite.ctx, folderPath, 0)
	suite.Require().NoError(err)
	suite.Equal(folderPath, lf.Metadata.Path)
}

func (suite *IntegrationTestSuite) Test_FolderOperations_ByID() {
	folderPath := suite.testFolderPath + "/go_pCloud_" + uuid.New().String()
	folderName := "go_pCloud_" + uuid.New().String()

	_, err := suite.pcc.DeleteFolderRecursive(suite.ctx, folderPath+"/"+folderName, 0)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	_, err = suite.pcc.DeleteFolder(suite.ctx, folderPath+"/"+folderName, 0)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d:", sdk.ErrDirectoryNotExists))

	lf, err := suite.pcc.CreateFolder(suite.ctx, "", suite.testFolderID, folderName)
	suite.Require().NoError(err)
	folderID := lf.Metadata.FolderID

	lf, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, "", suite.testFolderID, folderName)
	suite.Require().NoError(err)

	lf, err = suite.pcc.ListFolder(suite.ctx, "", folderID, true, false, false, false)
	suite.Require().NoError(err)

	fr, err := suite.pcc.DeleteFolderRecursive(suite.ctx, "", folderID)
	suite.Require().NoError(err)
	suite.EqualValues(1, fr.DeletedFolders)
	suite.EqualValues(0, fr.DeletedFiles)

	lf, err = suite.pcc.CreateFolderIfNotExists(suite.ctx, "", suite.testFolderID, folderName)
	suite.Require().NoError(err)
	folderID = lf.Metadata.FolderID

	lf, err = suite.pcc.DeleteFolder(suite.ctx, "", folderID)
	suite.Require().NoError(err)
	suite.EqualValues(folderID, lf.Metadata.FolderID)
}
