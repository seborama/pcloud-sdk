package sdk_test

import (
	"seborama/pcloud/sdk"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) Test_FileOps_ByPath() {
	folderPath := suite.testFolderPath + "/go_pCloud_" + uuid.New().String()
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	_, err := suite.pcc.CreateFolder(suite.ctx, folderPath, 0, "")
	suite.Require().NoError(err)

	// File operations by path
	f, err := suite.pcc.FileOpen(suite.ctx, sdk.O_CREAT|sdk.O_EXCL, folderPath+"/"+fileName, 0, 0, "")
	suite.Require().NoError(err)

	fdt, err := suite.pcc.FileWrite(suite.ctx, f.FD, []byte(Lipsum))
	suite.Require().NoError(err)
	suite.Require().EqualValues(len(Lipsum), fdt.Bytes)

	err = suite.pcc.FileClose(suite.ctx, f.FD)
	suite.Require().NoError(err)

	// copy original file to "* COPY", for use by "File operations by id", below
	cf, err := suite.pcc.CopyFile(suite.ctx, folderPath+"/"+fileName, 0, folderPath+"/"+fileName+" COPY", 0, "", true, time.Time{}, time.Time{})
	suite.Require().NoError(err)
	cFileID := cf.Metadata.FileID

	// copy original file to "* COPY2"
	cf2, err := suite.pcc.CopyFile(suite.ctx, folderPath+"/"+fileName, 0, folderPath+"/"+fileName+" COPY2", 0, "", true, time.Time{}, time.Time{})
	suite.Require().NoError(err)
	cFileID2 := cf2.Metadata.FileID

	// rename original file to "* COPY2" (i.e. overwrite operation)
	rf, err := suite.pcc.RenameFile(suite.ctx, folderPath+"/"+fileName, 0, folderPath+"/"+fileName+" COPY2", 0, "")
	suite.Require().NoError(err)
	suite.Equal(cFileID2, rf.Metadata.DeletedFileID)

	df, err := suite.pcc.DeleteFile(suite.ctx, folderPath+"/"+fileName+" COPY2", 0)
	suite.Require().NoError(err)
	suite.True(df.Metadata.IsDeleted)

	// File operations by id
	f, err = suite.pcc.FileOpen(suite.ctx, 0, "", cFileID, 0, "")
	suite.Require().NoError(err)

	err = suite.pcc.FileClose(suite.ctx, f.FD)
	suite.Require().NoError(err)

	rf, err = suite.pcc.RenameFile(suite.ctx, "", cFileID, "", suite.testFolderID, fileName+" RENAMED BY ID")
	suite.Require().NoError(err)

	df, err = suite.pcc.DeleteFile(suite.ctx, "", cFileID)
	suite.Require().NoError(err)
	suite.True(df.Metadata.IsDeleted)

	// _, err = suite.pcc.DeleteFolderRecursive(suite.ctx, folderPath, 0)
	// suite.Require().NoError(err)
}
