package sdk_test

import (
	"time"
)

func (suite *IntegrationTestSuite) Test_UserInfo() {
	ui, err := suite.pcc.UserInfo(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(ui.APIServer)
	suite.Require().NotEmpty(ui.Email)
}

func (suite *IntegrationTestSuite) Test_Diff() {
	dr, err := suite.pcc.Diff(suite.ctx, 0, time.Now().Add(-10*time.Minute), 0, false, 0)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(dr.DiffID, uint64(1))
	suite.Require().GreaterOrEqual(dr.Entries[0].DiffID, uint64(1))
	suite.Require().NotEmpty(dr.Entries[0].Metadata.Name)
}
