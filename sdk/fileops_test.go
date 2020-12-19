package sdk_test

import (
	"crypto/sha1"
	"fmt"
	"math"
	"seborama/pcloud/sdk"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) Test_FileOps_ByPath() {
	folderPath := suite.testFolderPath + "/go_pCloud_" + uuid.New().String()
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	_, err := suite.pcc.CreateFolder(suite.ctx, sdk.T2FolderByPath(folderPath))
	suite.Require().NoError(err)

	// File operations by path
	f, err := suite.pcc.FileOpen(suite.ctx, sdk.O_CREAT|sdk.O_EXCL, sdk.T4FileByPath(folderPath+"/"+fileName))
	suite.Require().NoError(err)

	fdt, err := suite.pcc.FileWrite(suite.ctx, f.FD, []byte(Lipsum))
	suite.Require().NoError(err)
	suite.Require().EqualValues(len(Lipsum), fdt.Bytes)

	fs, err := suite.pcc.FileSeek(suite.ctx, f.FD, 0, sdk.WhenceFromBeginning)
	suite.Require().NoError(err)
	suite.Require().Zero(fs.Offset)

	data, err := suite.pcc.FileRead(suite.ctx, f.FD, math.MaxInt64)
	suite.Require().NoError(err)
	suite.Require().EqualValues(Lipsum, data)

	dataPartial, err := suite.pcc.FilePRead(suite.ctx, f.FD, 20, 1000)
	suite.Require().NoError(err)
	suite.Require().EqualValues(Lipsum[1000:1020], string(dataPartial))

	cs := sha1.New()
	cs.Write(dataPartial)
	sha1sum := fmt.Sprintf("%x", cs.Sum(nil))
	dataPartial, err = suite.pcc.FilePReadIfMod(suite.ctx, f.FD, 20, 1000, sdk.T5SHA1(sha1sum))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), fmt.Sprintf("error %d: ", sdk.ErrNotModified))
	suite.Require().Empty(dataPartial)

	err = suite.pcc.FileClose(suite.ctx, f.FD)
	suite.Require().NoError(err)

	// copy original file to "* COPY", for use by "File operations by id", below
	cf, err := suite.pcc.CopyFile(suite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName), sdk.ToT3ByPath(folderPath+"/"+fileName+" COPY"), true, time.Time{}, time.Time{})
	suite.Require().NoError(err)
	cFileID := cf.Metadata.FileID

	// copy original file to "* COPY2"
	cf2, err := suite.pcc.CopyFile(suite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName), sdk.ToT3ByPath(folderPath+"/"+fileName+" COPY2"), true, time.Time{}, time.Time{})
	suite.Require().NoError(err)
	cFileID2 := cf2.Metadata.FileID

	// rename original file to "* COPY2" (i.e. overwrite operation)
	rf, err := suite.pcc.RenameFile(suite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName), sdk.ToT3ByPath(folderPath+"/"+fileName+" COPY2"))
	suite.Require().NoError(err)
	suite.Equal(cFileID2, rf.Metadata.DeletedFileID)

	// delete "* COPY2" file.
	// sometimes RenameFile() can be slightly delayed.
	// if this test fails, it is probably enough to just run it again.
	time.Sleep(500 * time.Millisecond)
	df, err := suite.pcc.DeleteFile(suite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName+" COPY2"))
	suite.Require().NoError(err)
	suite.True(df.Metadata.IsDeleted)

	// File operations by id
	f, err = suite.pcc.FileOpen(suite.ctx, 0, sdk.T4FileByID(cFileID))
	suite.Require().NoError(err)

	err = suite.pcc.FileClose(suite.ctx, f.FD)
	suite.Require().NoError(err)

	rf, err = suite.pcc.RenameFile(suite.ctx, sdk.T3FileByID(cFileID), sdk.ToT3ByIDName(suite.testFolderID, fileName+" RENAMED BY ID"))
	suite.Require().NoError(err)

	df, err = suite.pcc.DeleteFile(suite.ctx, sdk.T3FileByID(cFileID))
	suite.Require().NoError(err)
	suite.True(df.Metadata.IsDeleted)

	_, err = suite.pcc.DeleteFolderRecursive(suite.ctx, sdk.T1FolderByPath(folderPath))
	suite.Require().NoError(err)
}
