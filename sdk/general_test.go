package sdk_test

import (
	"time"
)

func (testsuite *IntegrationTestSuite) Test_UserInfo() {
	ui, err := testsuite.pcc.UserInfo(testsuite.ctx)
	testsuite.Require().NoError(err)
	testsuite.Require().NotEmpty(ui.APIServer)
	testsuite.Require().NotEmpty(ui.Email)
}

func (testsuite *IntegrationTestSuite) Test_GetFileHistory() {
	dr, err := testsuite.pcc.GetFileHistory(testsuite.ctx, testsuite.testFileID)
	testsuite.Require().NoError(err)
	testsuite.Require().GreaterOrEqual(dr.Entries[0].DiffID, uint64(1))
	testsuite.Require().NotEmpty(dr.Entries[0].Metadata.Name)
	testsuite.Require().EqualValues(testsuite.testFileID, dr.Entries[0].Metadata.FileID)
	testsuite.Require().EqualValues(testsuite.testFolderID, dr.Entries[0].Metadata.ParentFolderID)
}

func (testsuite *IntegrationTestSuite) Test_Diff() {
	dr, err := testsuite.pcc.Diff(testsuite.ctx, 0, time.Now().Add(-10*time.Minute), 0, false, 0)
	testsuite.Require().NoError(err)
	testsuite.Require().GreaterOrEqual(dr.DiffID, uint64(1))
	testsuite.Require().GreaterOrEqual(dr.Entries[0].DiffID, uint64(1))
	testsuite.Require().NotEmpty(dr.Entries[0].Metadata.Name)
}
