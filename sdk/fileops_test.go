package sdk_test

import (
	"crypto/sha1"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/seborama/pcloud-sdk/sdk"
)

func (testsuite *IntegrationTestSuite) Test_FileOps_ByPath() {
	folderPath := testsuite.testFolderPath + "/go_pCloud_" + uuid.New().String()
	fileName := "go_pCloud_" + uuid.New().String() + ".bin"

	_, err := testsuite.pcc.CreateFolder(testsuite.ctx, sdk.T2FolderByPath(folderPath))
	testsuite.Require().NoError(err)

	// File operations by path
	f, err := testsuite.pcc.FileOpen(testsuite.ctx, sdk.O_CREAT|sdk.O_EXCL, sdk.T4FileByPath(folderPath+"/"+fileName))
	testsuite.Require().NoError(err)

	// file write
	fdt, err := testsuite.pcc.FileWrite(testsuite.ctx, f.FD, []byte(Lipsum))
	testsuite.Require().NoError(err)
	testsuite.Require().EqualValues(len(Lipsum), fdt.Bytes)

	// file offset seek
	fs, err := testsuite.pcc.FileSeek(testsuite.ctx, f.FD, 0, sdk.WhenceFromBeginning)
	testsuite.Require().NoError(err)
	testsuite.Require().Zero(fs.Offset)

	// file read
	data, err := testsuite.pcc.FileRead(testsuite.ctx, f.FD, math.MaxInt64)
	testsuite.Require().NoError(err)
	testsuite.Require().EqualValues(Lipsum, data)

	// partial file read
	count := uint64(3200)
	offset := uint64(0)
	dataPartial, err := testsuite.pcc.FilePRead(testsuite.ctx, f.FD, count, offset)
	testsuite.Require().NoError(err)
	testsuite.Require().EqualValues(Lipsum[offset:(offset+count)], string(dataPartial))

	// conditional partial file read
	// nolint: gosec
	cs := sha1.New()
	_, err = cs.Write(dataPartial)
	testsuite.Require().NoError(err)

	sha1sum := fmt.Sprintf("%x", cs.Sum(nil))
	dataPartial, err = testsuite.pcc.FilePReadIfMod(testsuite.ctx, f.FD, count, offset, sdk.T5SHA1(sha1sum))
	testsuite.Require().Error(err)
	testsuite.Require().Contains(err.Error(), fmt.Sprintf("error %d: ", sdk.ErrNotModified))
	testsuite.Require().Empty(dataPartial)

	// partial file checksum
	pfc, err := testsuite.pcc.FileChecksum(testsuite.ctx, f.FD, count, offset)
	testsuite.Require().NoError(err)
	testsuite.EqualValues(sha1sum, pfc.SHA1)
	testsuite.EqualValues(pfc.Size, count)

	err = testsuite.pcc.FileClose(testsuite.ctx, f.FD)
	testsuite.Require().NoError(err)

	// full file checksum - with retry to allow pCloud to sync the changes made so far.
	for {
		time.Sleep(500 * time.Millisecond)
		cs.Reset()

		_, err = cs.Write([]byte(Lipsum))
		testsuite.Require().NoError(err)

		sha1sum = fmt.Sprintf("%x", cs.Sum(nil))
		var fc *sdk.FileChecksum
		fc, err = testsuite.pcc.ChecksumFile(testsuite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName))
		if err != nil {
			continue
		}
		if uint64(len(Lipsum)) != fc.Metadata.Size {
			// file changes have not quite yet fully propagated internally in pCloud
			continue
		}
		testsuite.EqualValues(sha1sum, fc.SHA1)
		testsuite.EqualValues(f.FileID, fc.Metadata.FileID)
		break
	}

	// copy original file to "* COPY", for use by "File operations by id", below
	cf, err := testsuite.pcc.CopyFile(testsuite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName), sdk.ToT3ByPath(folderPath+"/"+fileName+" COPY"), true, time.Time{}, time.Time{})
	testsuite.Require().NoError(err)
	cFileID := cf.Metadata.FileID

	// copy original file to "* COPY2"
	cf2, err := testsuite.pcc.CopyFile(testsuite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName), sdk.ToT3ByPath(folderPath+"/"+fileName+" COPY2"), true, time.Time{}, time.Time{})
	testsuite.Require().NoError(err)
	cFileID2 := cf2.Metadata.FileID

	// rename original file to "* COPY2" (i.e. overwrite operation)
	rf, err := testsuite.pcc.RenameFile(testsuite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName), sdk.ToT3ByPath(folderPath+"/"+fileName+" COPY2"))
	testsuite.Require().NoError(err)
	testsuite.Equal(cFileID2, rf.Metadata.DeletedFileID)

	// delete "* COPY2" file.
	df, err := testsuite.pcc.DeleteFile(testsuite.ctx, sdk.T3FileByPath(folderPath+"/"+fileName+" COPY2"))
	testsuite.Require().NoError(err)
	testsuite.True(df.Metadata.IsDeleted)

	// File operations by id
	f, err = testsuite.pcc.FileOpen(testsuite.ctx, 0, sdk.T4FileByID(cFileID))
	testsuite.Require().NoError(err)

	err = testsuite.pcc.FileClose(testsuite.ctx, f.FD)
	testsuite.Require().NoError(err)

	_, err = testsuite.pcc.RenameFile(testsuite.ctx, sdk.T3FileByID(cFileID), sdk.ToT3ByIDName(testsuite.testFolderID, fileName+" RENAMED BY ID"))
	testsuite.Require().NoError(err)

	df, err = testsuite.pcc.DeleteFile(testsuite.ctx, sdk.T3FileByID(cFileID))
	testsuite.Require().NoError(err)
	testsuite.True(df.Metadata.IsDeleted)

	_, err = testsuite.pcc.DeleteFolderRecursive(testsuite.ctx, sdk.T1FolderByPath(folderPath))
	testsuite.Require().NoError(err)
}
