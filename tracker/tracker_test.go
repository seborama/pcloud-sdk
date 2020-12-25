package tracker_test

import (
	"context"
	"io/ioutil"
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
	time7 := time.Now().Add(-18 * time.Hour)

	lf := pCloudFolderTreeSample1(time1, time2, time3, time4, time5, time6, time7)

	suite.pCloudClient.
		On("ListFolder", suite.ctx, mock.AnythingOfType("sdk.T1PathOrFolderID"), true, true, false, false, []sdk.ClientOption(nil)).
		Return(lf, nil).
		Once()

	err := suite.tracker.ListLatestPCloudContents(suite.ctx)
	suite.Require().NoError(err)

	expected := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	fsEntries, err := suite.store.GetLatestFileSystemEntries(suite.ctx, db.PCloudFileSystem)
	suite.Require().NoError(err)

	sortedEntries := func(elements []db.FSEntry) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedEntries(expected))
	sort.Slice(fsEntries, sortedEntries(fsEntries))
	suite.Equal("", cmp.Diff(expected, fsEntries, cmpopts.IgnoreUnexported()))
}

func (suite *IntegrationTestSuite) TestFindPCloudMutations_FilesDeleted() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh := suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			entriesCh <- e
		}
	}()

	err := <-errCh
	suite.Require().NoError(err)

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        0,
				IsFolder:       true,
				IsDeleted:      false,
				DeletedFileID:  0,
				Path:           "/",
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
				EntryID:        20001,
				IsFolder:       true,
				IsDeleted:      false,
				Path:           "/",
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
				IsDeleted:      false,
				Path:           "/Folder2",
				Name:           "File2",
				ParentFolderID: 20001,
				Created:        time5,
				Modified:       time5,
				Size:           789,
				Hash:           "9876543210100020002",
			},
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        30001,
				IsFolder:       true,
				IsDeleted:      false,
				Path:           "/",
				Name:           "Folder3",
				ParentFolderID: 0,
				Created:        time6,
				Modified:       time6,
			},
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        1000003,
				IsFolder:       false,
				IsDeleted:      false,
				Path:           "/",
				Name:           "File000",
				ParentFolderID: 0,
				Created:        time7,
				Modified:       time7,
				Size:           456,
				Hash:           "9876543210101000003",
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

func (suite *IntegrationTestSuite) TestFindPCloudMutations_FilesCreated() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh := suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			entriesCh <- e
		}
	}()

	err := <-errCh
	suite.Require().NoError(err)

	err = suite.store.MarkNewFileSystemEntriesAsPrevious(suite.ctx, db.PCloudFileSystem)

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        0,
				IsFolder:       true,
				IsDeleted:      false,
				DeletedFileID:  0,
				Path:           "/",
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
				EntryID:        20001,
				IsFolder:       true,
				IsDeleted:      false,
				Path:           "/",
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
				IsDeleted:      false,
				Path:           "/Folder2",
				Name:           "File2",
				ParentFolderID: 20001,
				Created:        time5,
				Modified:       time5,
				Size:           789,
				Hash:           "9876543210100020002",
			},
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        30001,
				IsFolder:       true,
				IsDeleted:      false,
				Path:           "/",
				Name:           "Folder3",
				ParentFolderID: 0,
				Created:        time6,
				Modified:       time6,
			},
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				EntryID:        1000003,
				IsFolder:       false,
				IsDeleted:      false,
				Path:           "/",
				Name:           "File000",
				ParentFolderID: 0,
				Created:        time7,
				Modified:       time7,
				Size:           456,
				Hash:           "9876543210101000003",
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

func (suite *IntegrationTestSuite) TestFindPCloudMutations_FileModified() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh := suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			entriesCh <- e
		}
	}()

	err := <-errCh
	suite.Require().NoError(err)

	err = suite.store.MarkNewFileSystemEntriesAsPrevious(suite.ctx, db.PCloudFileSystem)

	fse1 = fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh = suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			if e.EntryID == 20002 {
				e.Hash = "1234565432100020002"
			}
			entriesCh <- e
		}
	}()

	err = <-errCh
	suite.Require().NoError(err)

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeModified,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        20002,
				IsFolder:       false,
				IsDeleted:      false,
				Path:           "/Folder2",
				Name:           "File2",
				ParentFolderID: 20001,
				Created:        time5,
				Modified:       time5,
				Size:           789,
				Hash:           "1234565432100020002",
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

func (suite *IntegrationTestSuite) TestFindPCloudMutations_FileMoved() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh := suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			entriesCh <- e
		}
	}()

	err := <-errCh
	suite.Require().NoError(err)

	err = suite.store.MarkNewFileSystemEntriesAsPrevious(suite.ctx, db.PCloudFileSystem)

	fse1 = fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh = suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			if e.EntryID == 20002 {
				// move File2 to Folder3
				e.Path = "/Folder3"
				e.ParentFolderID = 30001
			}
			entriesCh <- e
		}
	}()

	err = <-errCh
	suite.Require().NoError(err)

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeMoved,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				EntryID:        20002,
				IsFolder:       false,
				IsDeleted:      false,
				Path:           "/Folder3",
				Name:           "File2",
				ParentFolderID: 30001,
				Created:        time5,
				Modified:       time5,
				Size:           789,
				Hash:           "9876543210100020002",
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

func (suite *IntegrationTestSuite) TestFindPCloudMutations_NoChanges() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh := suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			entriesCh <- e
		}
	}()

	err := <-errCh
	suite.Require().NoError(err)

	err = suite.store.MarkNewFileSystemEntriesAsPrevious(suite.ctx, db.PCloudFileSystem)

	fse1 = fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	entriesCh, errCh = suite.store.AddNewFileSystemEntry(suite.ctx, db.PCloudFileSystem)

	func() {
		defer close(entriesCh)
		for _, e := range fse1 {
			entriesCh <- e
		}
	}()

	err = <-errCh
	suite.Require().NoError(err)

	expected := []db.FSMutation{}

	fsMutations, err := suite.tracker.FindPCloudMutations(suite.ctx)
	suite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	suite.Equal("", cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()))
}

func (suite *IntegrationTestSuite) TestListLatestLocalContents() {
	now := time.Now()

	err := os.MkdirAll("./data_test/local", 0x750)
	suite.Require().NoError(err)
	err = ioutil.WriteFile("./data_test/local/File000", []byte("This is File000"), 0x750)
	suite.Require().NoError(err)

	err = os.MkdirAll("./data_test/local/Folder1", 0x750)
	suite.Require().NoError(err)
	err = ioutil.WriteFile("./data_test/local/Folder1/File1", []byte("This is File1"), 0x750)
	suite.Require().NoError(err)

	err = os.MkdirAll("./data_test/local/Folder2", 0x750)
	suite.Require().NoError(err)
	err = ioutil.WriteFile("./data_test/local/Folder2/File2", []byte("This is File2"), 0x750)
	suite.Require().NoError(err)

	err = os.MkdirAll("./data_test/local/Folder3", 0x750)
	suite.Require().NoError(err)
	err = ioutil.WriteFile("./data_test/local/Folder3/File3", []byte("This is File3"), 0x750)
	suite.Require().NoError(err)

	err = suite.tracker.ListLatestLocalContents(suite.ctx, "./data_test/local")
	suite.Require().NoError(err)

	expected := []db.FSEntry{
		{
			IsFolder:  true,
			IsDeleted: false,
			Path:      "data_test",
			Name:      "local",
			Hash:      "",
		},
		{
			IsFolder:  false,
			IsDeleted: false,
			Path:      "data_test/local",
			Name:      "File000",
			Size:      15,
			Hash:      "01ce643e7c1ca98f6fb21e61b5d03f547813edae",
		},
		{
			IsFolder:  true,
			IsDeleted: false,
			Path:      "data_test/local",
			Name:      "Folder1",
			Hash:      "",
		},
		{
			IsFolder:  false,
			IsDeleted: false,
			Path:      "data_test/local/Folder1",
			Name:      "File1",
			Size:      13,
			Hash:      "e8dfb879ddc708ea337a00e9b5580b498193bd2d",
		},
		{
			IsFolder:  true,
			IsDeleted: false,
			Path:      "data_test/local",
			Name:      "Folder2",
			Hash:      "",
		},
		{
			IsFolder:  false,
			IsDeleted: false,
			Path:      "data_test/local/Folder2",
			Name:      "File2",
			Size:      13,
			Hash:      "c28739a884e3742ea784f63dd52d9a4a90372235",
		},
		{
			IsFolder:  true,
			IsDeleted: false,
			Path:      "data_test/local",
			Name:      "Folder3",
			Hash:      "",
		},
		{
			IsFolder:  false,
			IsDeleted: false,
			Path:      "data_test/local/Folder3",
			Name:      "File3",
			Size:      13,
			Hash:      "995ea3c9a13945c2415a0881cc1bdac7a526b681",
		},
	}

	fsEntries, err := suite.store.GetLatestFileSystemEntries(suite.ctx, db.LocalFileSystem)
	suite.Require().NoError(err)
	suite.Require().Len(fsEntries, len(expected))

	sortedEntries := func(elements []db.FSEntry) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}

	sort.Slice(expected, sortedEntries(expected))
	sort.Slice(fsEntries, sortedEntries(fsEntries))

	for i, e := range expected {
		actualE := fsEntries[i]
		suite.NotEmpty(actualE.DeviceID)
		suite.Greater(actualE.EntryID, uint64(0))
		suite.EqualValues(e.IsFolder, actualE.IsFolder)
		suite.False(actualE.IsDeleted)
		suite.EqualValues(e.Path, actualE.Path, "for: "+actualE.Name)
		suite.EqualValues(e.Name, actualE.Name)
		suite.Greaterf(actualE.ParentFolderID, uint64(0), "expected: %s - actual: %s", e.Name, fsEntries[i].Name)
		suite.WithinDuration(now, actualE.Created, 5*time.Minute)
		suite.WithinDuration(now, actualE.Modified, 5*time.Minute)
		if !e.IsFolder {
			suite.EqualValues(e.Size, actualE.Size)
		}
		suite.EqualValues(e.Hash, actualE.Hash)
	}
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
// ├── Folder3
// └── File000
func pCloudFolderTreeSample1(time1, time2, time3, time4, time5, time6, time7 time.Time) *sdk.FSList {
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
					Contents:       []sdk.Metadata{},
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

// fsEntrySample1 counterpart to folderTreeSample1().
func fsEntrySample1(time1, time2, time3, time4, time5, time6, time7 time.Time) []db.FSEntry {
	return []db.FSEntry{
		{
			EntryID:        0,
			IsFolder:       true,
			IsDeleted:      false,
			DeletedFileID:  0,
			Path:           "/",
			Name:           "/",
			ParentFolderID: 0,
			Created:        time1,
			Modified:       time1,
		},
		{
			EntryID:        10001,
			IsFolder:       true,
			IsDeleted:      true,
			Path:           "/",
			Name:           "Folder1",
			ParentFolderID: 0,
			Created:        time2,
			Modified:       time2,
		},
		{
			EntryID:        10002,
			IsFolder:       false,
			IsDeleted:      true,
			Path:           "/Folder1",
			Name:           "File1",
			ParentFolderID: 10001,
			Created:        time3,
			Modified:       time3,
			Size:           123,
			Hash:           "9876543210123456789",
		},
		{
			EntryID:        20001,
			IsFolder:       true,
			IsDeleted:      false,
			Path:           "/",
			Name:           "Folder2",
			ParentFolderID: 0,
			Created:        time4,
			Modified:       time4,
		},
		{
			EntryID:        20002,
			IsFolder:       false,
			IsDeleted:      false,
			Path:           "/Folder2",
			Name:           "File2",
			ParentFolderID: 20001,
			Created:        time5,
			Modified:       time5,
			Size:           789,
			Hash:           "9876543210100020002",
		},
		{
			EntryID:        30001,
			IsFolder:       true,
			IsDeleted:      false,
			Path:           "/",
			Name:           "Folder3",
			ParentFolderID: 0,
			Created:        time6,
			Modified:       time6,
		},
		{
			EntryID:        1000003,
			IsFolder:       false,
			IsDeleted:      false,
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
