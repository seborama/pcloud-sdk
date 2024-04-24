package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/seborama/pcloud-sdk/sdk"
)

type IntegrationTestSuite struct {
	suite.Suite
	pcc *sdk.Client
	ctx context.Context

	testFolderPath string
	testFolderID   uint64
	testFileID     uint64
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (testsuite *IntegrationTestSuite) SetupSuite() {
	testsuite.ctx = context.Background()

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

	testsuite.initAuthenticatedClient(c)
	testsuite.initSuiteTestFolder()
}

func (testsuite *IntegrationTestSuite) TearDownSuite() {
	testsuite.deleteSuiteTestFolder()
	testsuite.logout()
}

func (testsuite *IntegrationTestSuite) initAuthenticatedClient(c *http.Client) {
	username := os.Getenv("GO_PCLOUD_USERNAME")
	testsuite.Require().NotEmpty(username)

	password := os.Getenv("GO_PCLOUD_PASSWORD")
	testsuite.Require().NotEmpty(password)

	otpCode := os.Getenv("GO_PCLOUD_TFA_CODE")

	pcc := sdk.NewClient(c)

	err := pcc.Login(
		testsuite.ctx,
		otpCode,
		sdk.WithGlobalOptionUsername(username),
		sdk.WithGlobalOptionPassword(password),
		sdk.WithGlobalOptionAuthInactiveExpire(5*time.Minute),
	)
	testsuite.Require().NoError(err)

	testsuite.pcc = pcc
}

func (testsuite *IntegrationTestSuite) initSuiteTestFolder() {
	testsuite.testFolderPath = "/goPCloudSDK_TestFolder_" + uuid.New().String()
	lf, err := testsuite.pcc.CreateFolder(testsuite.ctx, sdk.T2FolderByPath(testsuite.testFolderPath))
	testsuite.Require().NoError(err)
	testsuite.testFolderID = lf.Metadata.FolderID

	f, err := testsuite.pcc.FileOpen(testsuite.ctx, sdk.O_CREAT, sdk.T4FileByFolderIDName(testsuite.testFolderID, "sample.file"))
	testsuite.Require().NoError(err)
	testsuite.testFileID = f.FileID

	err = testsuite.pcc.FileClose(testsuite.ctx, f.FD)
	testsuite.Require().NoError(err)
}

func (testsuite *IntegrationTestSuite) logout() {
	lr, err := testsuite.pcc.Logout(testsuite.ctx)
	testsuite.Require().NoError(err)

	fmt.Println("auth token deleted:", lr.AuthDeleted)
}

func (testsuite *IntegrationTestSuite) deleteSuiteTestFolder() {
	_, err := testsuite.pcc.DeleteFolderRecursive(testsuite.ctx, sdk.T1FolderByID(testsuite.testFolderID))
	testsuite.NoError(err)
}
