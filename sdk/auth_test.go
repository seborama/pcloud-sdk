package sdk_test

func (suite *IntegrationTestSuite) Test_ListTokens() {
	_, err := suite.pcc.ListTokens(suite.ctx)
	suite.Require().NoError(err)
}
