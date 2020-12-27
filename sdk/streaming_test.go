package sdk_test

import (
	"seborama/pcloud/sdk"
	"time"

	"github.com/google/uuid"
)

func (testsuite *IntegrationTestSuite) Test_GetFileLink() {
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	f, err := testsuite.pcc.FileOpen(testsuite.ctx, sdk.O_CREAT|sdk.O_EXCL, sdk.T4FileByPath(testsuite.testFolderPath+"/"+fileName))
	testsuite.Require().NoError(err)

	fdt, err := testsuite.pcc.FileWrite(testsuite.ctx, f.FD, []byte(Lipsum))
	testsuite.Require().NoError(err)
	testsuite.Require().EqualValues(len(Lipsum), fdt.Bytes)

	err = testsuite.pcc.FileClose(testsuite.ctx, f.FD)
	testsuite.Require().NoError(err)

	fl, err := testsuite.pcc.GetFileLink(testsuite.ctx, sdk.T3FileByPath(testsuite.testFolderPath+"/"+fileName), true, "", 0, false)
	testsuite.Require().NoError(err)
	testsuite.Require().Equal(0, fl.Result)
	testsuite.Require().GreaterOrEqual(len(fl.Path), 10)
	testsuite.Require().EqualValues('/', fl.Path[0])
	testsuite.Require().True(fl.Expires.After(time.Now().Add(time.Hour)))
	testsuite.Require().GreaterOrEqual(len(fl.Hosts), 1)
}
