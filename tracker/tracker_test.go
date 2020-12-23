package tracker_test

import (
	"context"
	"os"
	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker"
	"seborama/pcloud/tracker/db"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	dbPath string
	store  *db.SQLite3

	pCloudClient *pCloudClientMock

	tracker *tracker.Tracker
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.dbPath = "./data_test"
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	_ = os.RemoveAll(suite.dbPath)
}

func (suite *IntegrationTestSuite) BeforeTest(suiteName, testName string) {
	suite.pCloudClient = &pCloudClientMock{}

	suite.makeDB()

	tracker, err := tracker.NewTracker(suite.ctx, suite.pCloudClient, suite.store)
	suite.Require().NoError(err)
	suite.tracker = tracker
}

func (suite *IntegrationTestSuite) makeDB() {
	const dbPath = "./data_test"

	err := os.RemoveAll(dbPath)
	suite.Require().NoError(err)

	err = os.Mkdir(dbPath, 0x770)
	suite.Require().NoError(err)

	dbase, err := db.NewSQLite3(suite.ctx, dbPath)
	suite.Require().NoError(err)

	suite.store = dbase
}

func (suite *IntegrationTestSuite) TestListLatestPCloudContents() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)

	lf := folderTreeSample1(time1, time2, time3, time4, time5, time6)

	suite.pCloudClient.
		On("ListFolder", suite.ctx, mock.AnythingOfType("sdk.T1PathOrFolderID"), true, true, false, false, []sdk.ClientOption(nil)).
		Return(lf, nil).
		Once()

	err := suite.tracker.ListLatestPCloudContents(suite.ctx)
	suite.Require().NoError(err)

	expected := fsEntrySample1(time1, time2, time3, time4, time5, time6)

	fsEntries, err := suite.store.GetLatestFileSystemEntries(suite.ctx)
	suite.Require().NoError(err)

	sortedEntries := func(elements []db.FSEntry) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedEntries(expected))
	sort.Slice(fsEntries, sortedEntries(fsEntries))
	suite.Equal("", cmp.Diff(expected, fsEntries, cmpopts.IgnoreUnexported()))
}

func (suite *IntegrationTestSuite) TestFindPCloudMutations_AllNewFiles() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6)
	for _, e := range fse1 {
		err := suite.store.AddNewFileSystemEntry(suite.ctx, e)
		suite.Require().NoError(err)
	}

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        0,
				IsFolder:       true,
				Deleted:        false,
				DeletedFileID:  0,
				Name:           "/",
				ParentFolderID: 0,
				Created:        time1,
				Modified:       time1,
			},
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        10001,
				IsFolder:       true,
				Deleted:        true,
				Name:           "Folder1",
				ParentFolderID: 0,
				Created:        time2,
				Modified:       time2,
			},
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        10002,
				IsFolder:       false,
				Deleted:        true,
				Name:           "File1",
				ParentFolderID: 10001,
				Created:        time3,
				Modified:       time3,
				Size:           123,
				Hash:           9876543210123456789,
			},
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        20001,
				IsFolder:       true,
				Deleted:        false,
				Name:           "Folder2",
				ParentFolderID: 0,
				Created:        time4,
				Modified:       time4,
			},
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        20002,
				IsFolder:       false,
				Deleted:        false,
				Name:           "File2",
				ParentFolderID: 20001,
				Created:        time5,
				Modified:       time5,
				Size:           789,
				Hash:           9876543210100020002,
			},
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        30003,
				IsFolder:       false,
				Deleted:        false,
				Name:           "File3",
				ParentFolderID: 0,
				Created:        time6,
				Modified:       time6,
				Size:           456,
				Hash:           9876543210100030003,
			},
		},
	}

	fsMutations, err := suite.tracker.FindPCloudMutations(suite.ctx)
	suite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	suite.Equal("", cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()))
}

func (suite *IntegrationTestSuite) TestFindPCloudMutations_AllOldFiles() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6)
	for _, e := range fse1 {
		err := suite.store.AddNewFileSystemEntry(suite.ctx, e)
		suite.Require().NoError(err)
	}

	err := suite.store.MarkNewFileSystemEntriesAsPrevious(suite.ctx)

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        0,
				IsFolder:       true,
				Deleted:        false,
				DeletedFileID:  0,
				Name:           "/",
				ParentFolderID: 0,
				Created:        time1,
				Modified:       time1,
			},
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        10001,
				IsFolder:       true,
				Deleted:        true,
				Name:           "Folder1",
				ParentFolderID: 0,
				Created:        time2,
				Modified:       time2,
			},
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        10002,
				IsFolder:       false,
				Deleted:        true,
				Name:           "File1",
				ParentFolderID: 10001,
				Created:        time3,
				Modified:       time3,
				Size:           123,
				Hash:           9876543210123456789,
			},
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        20001,
				IsFolder:       true,
				Deleted:        false,
				Name:           "Folder2",
				ParentFolderID: 0,
				Created:        time4,
				Modified:       time4,
			},
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        20002,
				IsFolder:       false,
				Deleted:        false,
				Name:           "File2",
				ParentFolderID: 20001,
				Created:        time5,
				Modified:       time5,
				Size:           789,
				Hash:           9876543210100020002,
			},
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        30003,
				IsFolder:       false,
				Deleted:        false,
				Name:           "File3",
				ParentFolderID: 0,
				Created:        time6,
				Modified:       time6,
				Size:           456,
				Hash:           9876543210100030003,
			},
		},
	}

	fsMutations, err := suite.tracker.FindPCloudMutations(suite.ctx)
	suite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	suite.Equal("", cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()))
}

type pCloudClientMock struct {
	mock.Mock
}

func (m *pCloudClientMock) ListFolder(ctx context.Context, folder sdk.T1PathOrFolderID, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...sdk.ClientOption) (*sdk.FSList, error) {
	args := m.Called(ctx, folder, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt, opts)
	return args.Get(0).(*sdk.FSList), args.Error(1)
}

func (m *pCloudClientMock) Diff(ctx context.Context, diffID uint64, after time.Time, last uint64, block bool, limit uint64, opts ...sdk.ClientOption) (*sdk.DiffResult, error) {
	args := m.Called(ctx, diffID, after, last, block, limit, opts)
	return args.Get(0).(*sdk.DiffResult), args.Error(1)
}

// /
// ├── Folder1 (deleted)
// │   ├── File1 (deleted)
// ├── Folder2
// │   ├── File2
// └── File3
func folderTreeSample1(time1, time2, time3, time4, time5, time6 time.Time) *sdk.FSList {
	return &sdk.FSList{
		Metadata: sdk.Metadata{
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
			Contents: []sdk.Metadata{
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
					Contents: []sdk.Metadata{
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
					Contents: []sdk.Metadata{
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
					Name: "File3",
					Created: &sdk.APITime{
						Time: time6,
					},
					IsMine: true,
					Thumb:  false,
					Modified: &sdk.APITime{
						Time: time6,
					},
					Comments:       0,
					ID:             "f30003",
					IsShared:       false,
					Icon:           "file",
					IsFolder:       false,
					ParentFolderID: 0,
					IsDeleted:      false,
					FileID:         30003,
					Hash:           9876543210100030003,
					Size:           456,
					ContentType:    "application/octet-stream",
				},
			},
		},
	}
}

// fsEntrySample1 counterpart to folderTreeSample1().
func fsEntrySample1(time1, time2, time3, time4, time5, time6 time.Time) []db.FSEntry {
	return []db.FSEntry{
		{
			EntryID:        0,
			IsFolder:       true,
			Deleted:        false,
			DeletedFileID:  0,
			Name:           "/",
			ParentFolderID: 0,
			Created:        time1,
			Modified:       time1,
		},
		{
			EntryID:        10001,
			IsFolder:       true,
			Deleted:        true,
			Name:           "Folder1",
			ParentFolderID: 0,
			Created:        time2,
			Modified:       time2,
		},
		{
			EntryID:        10002,
			IsFolder:       false,
			Deleted:        true,
			Name:           "File1",
			ParentFolderID: 10001,
			Created:        time3,
			Modified:       time3,
			Size:           123,
			Hash:           9876543210123456789,
		},
		{
			EntryID:        20001,
			IsFolder:       true,
			Deleted:        false,
			Name:           "Folder2",
			ParentFolderID: 0,
			Created:        time4,
			Modified:       time4,
		},
		{
			EntryID:        20002,
			IsFolder:       false,
			Deleted:        false,
			Name:           "File2",
			ParentFolderID: 20001,
			Created:        time5,
			Modified:       time5,
			Size:           789,
			Hash:           9876543210100020002,
		},
		{
			EntryID:        30003,
			IsFolder:       false,
			Deleted:        false,
			Name:           "File3",
			ParentFolderID: 0,
			Created:        time6,
			Modified:       time6,
			Size:           456,
			Hash:           9876543210100030003,
		},
	}
}
