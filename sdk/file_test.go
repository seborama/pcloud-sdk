package sdk_test

import (
	"os"
	"seborama/pcloud/sdk"
	"time"

	"github.com/google/uuid"
)

func (testsuite *IntegrationTestSuite) Test_UploadFile() {
	files := testsuite.createFiles()
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
	fu, err := testsuite.pcc.UploadFile(testsuite.ctx, sdk.T1FolderByID(testsuite.testFolderID), files, true, progressHash, true, time.Time{}, time.Time{})
	// if this test starts failing for no apparent reason, add a retry loop to ensure pCloud has propagated the upload(s).
	testsuite.Require().NoError(err)
	testsuite.Len(fu.FileIDs, len(files))
	testsuite.Len(fu.Checksums, len(files))
	for _, m := range fu.Metadata {
		testsuite.EqualValues(testsuite.testFolderID, m.ParentFolderID)
		fi, err := os.Stat(files[m.Name].Name())
		testsuite.Require().NoError(err)
		testsuite.EqualValues(fi.Size(), m.Size)
	}
}

func (testsuite *IntegrationTestSuite) createFiles() map[string]*os.File {
	num := 3
	files := map[string]*os.File{}

	for i := 0; i < num; i++ {
		fName := "Test_UploadFile_" + uuid.New().String()

		f, err := os.Create("/tmp/" + fName)
		testsuite.Require().NoError(err)

		_, err = f.WriteString("data for this file: " + uuid.New().String())
		testsuite.Require().NoError(err)

		_, err = f.Seek(0, 0) // note: the behaviour of Seek on a file opened with O_APPEND is not specified.
		testsuite.Require().NoError(err)
		files[fName] = f
	}

	return files
}
