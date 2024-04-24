package filesystem_test

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/seborama/pcloud-sdk/sdk"
	"github.com/seborama/pcloud-sdk/tracker/db"
	"github.com/seborama/pcloud-sdk/tracker/filesystem"
)

type PCloudIntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	dbPath string

	pCloudClient *pCloudClientMock

	pcloudFS *filesystem.PCloud
}

func TestPCloudIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PCloudIntegrationTestSuite))
}

func (testsuite *PCloudIntegrationTestSuite) SetupSuite() {
	testsuite.ctx = context.Background()
	testsuite.dbPath = "/tmp/data_test"
}

func (testsuite *PCloudIntegrationTestSuite) TearDownSuite() {
	_ = os.RemoveAll(testsuite.dbPath)
}

func (testsuite *PCloudIntegrationTestSuite) BeforeTest(suiteName, testName string) {
	testsuite.pCloudClient = &pCloudClientMock{}

	testsuite.makeDB()

	pcloudFS := filesystem.NewPCloud(testsuite.pCloudClient)
	testsuite.pcloudFS = pcloudFS
}

func (testsuite *PCloudIntegrationTestSuite) AfterTest(suiteName, testName string) {
	testsuite.pCloudClient.AssertExpectations(testsuite.T())
}

func (testsuite *PCloudIntegrationTestSuite) makeDB() {
	err := os.RemoveAll(testsuite.dbPath)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(testsuite.dbPath, 0700)
	testsuite.Require().NoError(err)
}

func (testsuite *PCloudIntegrationTestSuite) TestPCloud_Walk() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	lf := pCloudFolderTreeSample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.pCloudClient.
		On("ListFolder", testsuite.ctx, mock.AnythingOfType("sdk.T1PathOrFolderID"), true, false, false, false, []sdk.ClientOption(nil)).
		Return(lf, nil).
		Once()

	fsEntriesCh := make(chan db.FSEntry)
	errCh := make(chan error)
	fsEntries := []db.FSEntry{}

	go func() {
		for fse := range fsEntriesCh {
			fsEntries = append(fsEntries, fse)
		}

		errCh <- nil
	}()

	err := testsuite.pcloudFS.Walk(testsuite.ctx, "pcloud_fs", "/", fsEntriesCh, errCh)
	testsuite.Require().NoError(err)

	expected := fsEntrySample1(time1, time4, time5, time6, time7)

	sortedEntries := func(elements []db.FSEntry) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedEntries(expected))
	sort.Slice(fsEntries, sortedEntries(fsEntries))
	if d := cmp.Diff(expected, fsEntries, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

// /
// ├── Folder1 (deleted)
// │   ├── File1 (deleted)
// ├── Folder2
// │   ├── File2
// ├── Folder3
// └── File000
// nolint: dupl
func pCloudFolderTreeSample1(time1, time2, time3, time4, time5, time6, time7 time.Time) *sdk.FSList {
	return &sdk.FSList{
		Metadata: &sdk.Metadata{
			Path: "/",
			Name: "/",
			Created: &sdk.APITime{
				Time: time1,
			},
			IsMine: true,
			Thumb:  false,
			Modified: &sdk.APITime{
				Time: time1,
			},
			Comments:       0,
			ID:             "d0",
			IsShared:       false,
			Icon:           "folder",
			IsFolder:       true,
			ParentFolderID: 0,
			IsDeleted:      false,
			DeletedFileID:  0,
			FolderID:       0,
			Contents: []*sdk.Metadata{
				{
					Name: "Folder1",
					Created: &sdk.APITime{
						Time: time2,
					},
					IsMine: true,
					Thumb:  false,
					Modified: &sdk.APITime{
						Time: time2,
					},
					Comments:       0,
					ID:             "d10001",
					IsShared:       false,
					Icon:           "folder",
					IsFolder:       true,
					ParentFolderID: 0,
					IsDeleted:      true,
					FolderID:       10001,
					Contents: []*sdk.Metadata{
						{
							Name: "File1",
							Created: &sdk.APITime{
								Time: time3,
							},
							IsMine: true,
							Thumb:  false,
							Modified: &sdk.APITime{
								Time: time3,
							},
							Comments:       0,
							ID:             "f10002",
							IsShared:       false,
							Icon:           "file",
							IsFolder:       false,
							ParentFolderID: 10001,
							IsDeleted:      true,
							FileID:         10002,
							Hash:           9876543210123456789,
							Size:           123,
							ContentType:    "application/octet-stream",
						},
					},
				},
				{
					Name: "Folder2",
					Created: &sdk.APITime{
						Time: time4,
					},
					IsMine: true,
					Thumb:  false,
					Modified: &sdk.APITime{
						Time: time4,
					},
					Comments:       0,
					ID:             "d20001",
					IsShared:       false,
					Icon:           "folder",
					IsFolder:       true,
					ParentFolderID: 0,
					IsDeleted:      false,
					FolderID:       20001,
					Contents: []*sdk.Metadata{
						{
							Name: "File2",
							Created: &sdk.APITime{
								Time: time5,
							},
							IsMine: true,
							Thumb:  false,
							Modified: &sdk.APITime{
								Time: time5,
							},
							Comments:       0,
							ID:             "f20002",
							IsShared:       false,
							Icon:           "file",
							IsFolder:       false,
							ParentFolderID: 20001,
							IsDeleted:      false,
							FileID:         20002,
							Hash:           9876543210100020002,
							Size:           789,
							ContentType:    "application/octet-stream",
						},
					},
				},
				{
					Name: "Folder3",
					Created: &sdk.APITime{
						Time: time6,
					},
					IsMine: true,
					Thumb:  false,
					Modified: &sdk.APITime{
						Time: time6,
					},
					Comments:       0,
					ID:             "d30001",
					IsShared:       false,
					Icon:           "folder",
					IsFolder:       true,
					ParentFolderID: 0,
					IsDeleted:      false,
					FolderID:       30001,
					Contents:       []*sdk.Metadata{},
				},
				{
					Name: "File000",
					Created: &sdk.APITime{
						Time: time7,
					},
					IsMine: true,
					Thumb:  false,
					Modified: &sdk.APITime{
						Time: time7,
					},
					Comments:       0,
					ID:             "f1000003",
					IsShared:       false,
					Icon:           "file",
					IsFolder:       false,
					ParentFolderID: 0,
					IsDeleted:      false,
					FileID:         1000003,
					Hash:           9876543210101000003,
					Size:           456,
					ContentType:    "application/octet-stream",
				},
			},
		},
	}
}

type pCloudClientMock struct {
	mock.Mock
}

func (m *pCloudClientMock) ListFolder(ctx context.Context, folder sdk.T1PathOrFolderID, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...sdk.ClientOption) (*sdk.FSList, error) {
	args := m.Called(ctx, folder, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt, opts)
	return args.Get(0).(*sdk.FSList), args.Error(1)
}

// fsEntrySample1 counterpart to folderTreeSample1().
func fsEntrySample1(time1, time4, time5, time6, time7 time.Time) []db.FSEntry {
	return []db.FSEntry{
		{
			FSName:         "pcloud_fs",
			EntryID:        0,
			IsFolder:       true,
			Path:           "/",
			Name:           "/",
			ParentFolderID: 0,
			Created:        time1,
			Modified:       time1,
		},
		{
			FSName:         "pcloud_fs",
			EntryID:        20001,
			IsFolder:       true,
			Path:           "/",
			Name:           "Folder2",
			ParentFolderID: 0,
			Created:        time4,
			Modified:       time4,
		},
		{
			FSName:         "pcloud_fs",
			EntryID:        20002,
			IsFolder:       false,
			Path:           "/Folder2",
			Name:           "File2",
			ParentFolderID: 20001,
			Created:        time5,
			Modified:       time5,
			Size:           789,
			Hash:           "9876543210100020002",
		},
		{
			FSName:         "pcloud_fs",
			EntryID:        30001,
			IsFolder:       true,
			Path:           "/",
			Name:           "Folder3",
			ParentFolderID: 0,
			Created:        time6,
			Modified:       time6,
		},
		{
			FSName:         "pcloud_fs",
			EntryID:        1000003,
			IsFolder:       false,
			Path:           "/",
			Name:           "File000",
			ParentFolderID: 0,
			Created:        time7,
			Modified:       time7,
			Size:           456,
			Hash:           "9876543210101000003",
		},
	}
}
