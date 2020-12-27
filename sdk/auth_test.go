package sdk_test

func (testsuite *IntegrationTestSuite) Test_ListTokens() {
	_, err := testsuite.pcc.ListTokens(testsuite.ctx)
	testsuite.Require().NoError(err)
}
