package sdk_test

import (
	"seborama/pcloud/sdk"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) Test_DeleteFile_ByPath() {
	folderPath := suite.testFolderPath + "/go_pCloud_" + uuid.New().String()
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	pathFilename := folderPath + "/" + fileName

	_, err := suite.pcc.CreateFolder(suite.ctx, sdk.T2FolderByPath(folderPath))
	suite.Require().NoError(err)

	// FileOpenByPath
	f, err := suite.pcc.FileOpen(suite.ctx, sdk.O_CREAT|sdk.O_EXCL, pathFilename, 0, 0, "")
	suite.Require().NoError(err)
	fileID := f.FileID

	fdt, err := suite.pcc.FileWrite(suite.ctx, f.FD, []byte(Lipsum))
	suite.Require().NoError(err)
	suite.Require().EqualValues(len(Lipsum), fdt.Bytes)

	err = suite.pcc.FileClose(suite.ctx, f.FD)
	suite.Require().NoError(err)

	// FileOpenByID
	f, err = suite.pcc.FileOpen(suite.ctx, 0, "", fileID, 0, "")
	suite.Require().NoError(err)

	err = suite.pcc.FileClose(suite.ctx, f.FD)
	suite.Require().NoError(err)

	_, err = suite.pcc.DeleteFile(suite.ctx, pathFilename, 0)
	suite.Require().NoError(err)

	_, err = suite.pcc.DeleteFolderRecursive(suite.ctx, sdk.T1FolderByPath(folderPath))
	suite.Require().NoError(err)
}
