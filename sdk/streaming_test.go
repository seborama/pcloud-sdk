package sdk_test

import (
	"seborama/pcloud/sdk"
	"time"

	"github.com/google/uuid"
)

func (suite *IntegrationTestSuite) Test_GetFileLink() {
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	f, err := suite.pcc.FileOpen(suite.ctx, sdk.O_CREAT|sdk.O_EXCL, suite.testFolderPath+"/"+fileName, 0, 0, "")
	suite.Require().NoError(err)

	fdt, err := suite.pcc.FileWrite(suite.ctx, f.FD, []byte(Lipsum))
	suite.Require().NoError(err)
	suite.Require().EqualValues(len(Lipsum), fdt.Bytes)

	err = suite.pcc.FileClose(suite.ctx, f.FD)
	suite.Require().NoError(err)

	fl, err := suite.pcc.GetFileLink(suite.ctx, sdk.T3FileByPath(suite.testFolderPath+"/"+fileName), true, "", 0, false)
	suite.Require().NoError(err)
	suite.Require().Equal(0, fl.Result)
	suite.Require().GreaterOrEqual(len(fl.Path), 10)
	suite.Require().EqualValues('/', fl.Path[0])
	suite.Require().True(fl.Expires.After(time.Now().Add(time.Hour)))
	suite.Require().GreaterOrEqual(len(fl.Hosts), 1)
}
