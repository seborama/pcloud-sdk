package sdk_test

import (
	"os"
	"seborama/pcloud/sdk"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) Test_UploadFile() {
	files := suite.createFiles()
	defer func(files map[string]*os.File) {
		for _, f := range files {
			if f == nil {
				continue
			}
			fName := f.Name()
			f.Close()
			os.Remove(fName)
		}
	}(files)

	progressHash := ""
	fu, err := suite.pcc.UploadFile(suite.ctx, sdk.T1FolderByID(suite.testFolderID), files, true, progressHash, true, time.Time{}, time.Time{})
	// if this test starts failing for no apparent reason, add a retry loop to ensure pCloud has propagated the upload(s).
	suite.Require().NoError(err)
	suite.Len(fu.FileIDs, len(files))
	suite.Len(fu.Checksums, len(files))
	for _, m := range fu.Metadata {
		suite.EqualValues(suite.testFolderID, m.ParentFolderID)
		fi, err := os.Stat(files[m.Name].Name())
		suite.Require().NoError(err)
		suite.EqualValues(fi.Size(), m.Size)

	}
}

func (suite *IntegrationTestSuite) createFiles() map[string]*os.File {
	num := 3
	files := map[string]*os.File{}

	for i := 0; i < num; i++ {
		fName := "Test_UploadFile_" + uuid.New().String()

		f, err := os.Create("/tmp/" + fName)
		suite.Require().NoError(err)

		_, err = f.WriteString("data for this file: " + uuid.New().String())
		suite.Require().NoError(err)

		_, err = f.Seek(0, 0) // note: the behavior of Seek on a file opened with O_APPEND is not specified.
		suite.Require().NoError(err)
		files[fName] = f
	}

	return files
}
