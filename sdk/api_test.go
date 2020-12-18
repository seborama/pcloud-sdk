package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"seborama/pcloud/sdk"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	pcc *sdk.Client
	ctx context.Context

	testFolderPath string
	testFolderID   uint64
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	c := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       1,
			ResponseHeaderTimeout: 20 * time.Second,
			// Proxy:           http.ProxyFromEnvironment,
			// TLSClientConfig: &tls.Config{
			// InsecureSkipVerify: true, // only use this for debugging environments
			// },
		},
		Timeout: 0,
	}

	suite.initAuthenticatedClient(c)
	suite.initSuiteTestFolder()
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.deleteSuiteTestFolder()
	suite.logout()
}

// func (suite *IntegrationTestSuite) TestSDK() {
// 	suite.test_UserInfo()
// 	suite.test_Diff()
// }

func (suite *IntegrationTestSuite) initAuthenticatedClient(c *http.Client) {
	username := os.Getenv("GO_PCLOUD_USERNAME")
	suite.Require().NotEmpty(username)

	password := os.Getenv("GO_PCLOUD_PASSWORD")
	suite.Require().NotEmpty(password)

	otpCode := os.Getenv("GO_PCLOUD_TFA_CODE")

	pcc := sdk.NewClient(c)

	err := pcc.Login(
		suite.ctx,
		otpCode,
		sdk.WithGlobalOptionUsername(username),
		sdk.WithGlobalOptionPassword(password),
		sdk.WithGlobalOptionAuthInactiveExpire(5*time.Minute),
	)
	suite.Require().NoError(err)

	suite.pcc = pcc
}

func (suite *IntegrationTestSuite) initSuiteTestFolder() {
	suite.testFolderPath = "/goPCloudSDK_TestFolder_" + uuid.New().String()
	lf, err := suite.pcc.CreateFolder(suite.ctx, sdk.T2FolderByPath(suite.testFolderPath))
	suite.Require().NoError(err)
	suite.testFolderID = lf.Metadata.FolderID
}

func (suite *IntegrationTestSuite) logout() {
	lr, err := suite.pcc.Logout(suite.ctx)
	suite.NoError(err)

	fmt.Println("auth token deleted:", lr.AuthDeleted)
}

func (suite *IntegrationTestSuite) deleteSuiteTestFolder() {
	_, err := suite.pcc.DeleteFolderRecursive(suite.ctx, sdk.T1FolderByID(suite.testFolderID))
	suite.NoError(err)
}
