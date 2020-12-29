package tracker_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
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

func (testsuite *IntegrationTestSuite) SetupSuite() {
	testsuite.ctx = context.Background()
	testsuite.dbPath = "/tmp/data_test"
}

func (testsuite *IntegrationTestSuite) TearDownSuite() {
	_ = os.RemoveAll(testsuite.dbPath)
}

func (testsuite *IntegrationTestSuite) BeforeTest(suiteName, testName string) {
	testsuite.pCloudClient = &pCloudClientMock{}

	testsuite.makeDB()

	track, err := tracker.NewTracker(testsuite.ctx, testsuite.pCloudClient, testsuite.store)
	testsuite.Require().NoError(err)
	testsuite.tracker = track
}

func (testsuite *IntegrationTestSuite) AfterTest(suiteName, testName string) {
	err := testsuite.store.Close()
	testsuite.Require().NoError(err)
}

func (testsuite *IntegrationTestSuite) makeDB() {
	err := os.RemoveAll(testsuite.dbPath)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(testsuite.dbPath, 0700)
	testsuite.Require().NoError(err)

	dbase, err := db.NewSQLite3(testsuite.ctx, testsuite.dbPath)
	testsuite.Require().NoError(err)

	testsuite.store = dbase
}

func (testsuite *IntegrationTestSuite) TestListLatestPCloudContents() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	lf := pCloudFolderTreeSample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.pCloudClient.
		On("ListFolder", testsuite.ctx, mock.AnythingOfType("sdk.T1PathOrFolderID"), true, true, false, false, []sdk.ClientOption(nil)).
		Return(lf, nil).
		Once()

	err := testsuite.tracker.ListLatestPCloudContents(testsuite.ctx, tracker.WithEntriesChannelSize(0))
	testsuite.Require().NoError(err)

	expected := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	fsEntries, err := testsuite.store.GetLatestFileSystemEntries(testsuite.ctx, db.PCloudFileSystem)
	testsuite.Require().NoError(err)

	sortedEntries := func(elements []db.FSEntry) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedEntries(expected))
	sort.Slice(fsEntries, sortedEntries(fsEntries))
	if d := cmp.Diff(expected, fsEntries, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindPCloudMutations_FilesDeleted() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	// nolint: dupl
	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				EntryID:        0,
				IsFolder:       true,
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
				FSType:         db.PCloudFileSystem,
				EntryID:        20001,
				IsFolder:       true,
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
				FSType:         db.PCloudFileSystem,
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
		},
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				EntryID:        30001,
				IsFolder:       true,
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
				FSType:         db.PCloudFileSystem,
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
		},
	}

	fsMutations, err := testsuite.tracker.FindPCloudMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindPCloudMutations_FilesCreated() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.MarkNewFileSystemEntriesAsPrevious(testsuite.ctx, db.PCloudFileSystem)
	testsuite.Require().NoError(err)

	// nolint: dupl
	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				EntryID:        0,
				IsFolder:       true,
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
				FSType:         db.PCloudFileSystem,
				EntryID:        20001,
				IsFolder:       true,
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
				FSType:         db.PCloudFileSystem,
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
		},
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionPrevious,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				EntryID:        30001,
				IsFolder:       true,
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
				FSType:         db.PCloudFileSystem,
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
		},
	}

	fsMutations, err := testsuite.tracker.FindPCloudMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindPCloudMutations_FileModified() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.MarkNewFileSystemEntriesAsPrevious(testsuite.ctx, db.PCloudFileSystem)
	testsuite.Require().NoError(err)

	fse1 = fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, func(e db.FSEntry) db.FSEntry {
		if e.EntryID == 20002 {
			e.Hash = "1234565432100020002"
		}
		return e
	})

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeModified,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				EntryID:        20002,
				IsFolder:       false,
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

	fsMutations, err := testsuite.tracker.FindPCloudMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindPCloudMutations_FileMoved() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.MarkNewFileSystemEntriesAsPrevious(testsuite.ctx, db.PCloudFileSystem)
	testsuite.Require().NoError(err)

	fse1 = fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, func(e db.FSEntry) db.FSEntry {
		if e.EntryID == 20002 {
			// move File2 to Folder3
			e.Path = "/Folder3"
			e.ParentFolderID = 30001
		}
		return e
	})

	expected := []db.FSMutation{
		{
			Type:    db.MutationTypeMoved,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				EntryID:        20002,
				IsFolder:       false,
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

	fsMutations, err := testsuite.tracker.FindPCloudMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindPCloudMutations_NoChanges() {
	time1 := time.Now().Add(-24 * time.Hour)
	time2 := time.Now().Add(-23 * time.Hour)
	time3 := time.Now().Add(-22 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.MarkNewFileSystemEntriesAsPrevious(testsuite.ctx, db.PCloudFileSystem)
	testsuite.Require().NoError(err)

	fse1 = fsEntrySample1(time1, time2, time3, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	expected := []db.FSMutation{}

	fsMutations, err := testsuite.tracker.FindPCloudMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements []db.FSMutation) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestListLatestLocalContents() {
	now := time.Now()

	err := os.MkdirAll(filepath.Join(testsuite.dbPath, "local"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.dbPath, "local", "File000"), []byte("This is File000"), 0600)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(filepath.Join(testsuite.dbPath, "local", "Folder1"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.dbPath, "local", "Folder1", "File1"), []byte("This is File1"), 0600)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(filepath.Join(testsuite.dbPath, "local", "Folder2"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.dbPath, "local", "Folder2", "File2"), []byte("This is File2"), 0600)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(filepath.Join(testsuite.dbPath, "local", "Folder3"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.dbPath, "local", "Folder3", "File3"), []byte("This is File3"), 0600)
	testsuite.Require().NoError(err)

	err = testsuite.tracker.ListLatestLocalContents(testsuite.ctx, filepath.Join(testsuite.dbPath, "local"), tracker.WithEntriesChannelSize(0))
	testsuite.Require().NoError(err)

	expected := []db.FSEntry{
		{
			IsFolder: true,
			Path:     testsuite.dbPath,
			Name:     "local",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.dbPath, "local"),
			Name:     "File000",
			Size:     15,
			Hash:     "01ce643e7c1ca98f6fb21e61b5d03f547813edae",
		},
		{
			IsFolder: true,
			Path:     filepath.Join(testsuite.dbPath, "local"),
			Name:     "Folder1",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.dbPath, "local", "Folder1"),
			Name:     "File1",
			Size:     13,
			Hash:     "e8dfb879ddc708ea337a00e9b5580b498193bd2d",
		},
		{
			IsFolder: true,
			Path:     filepath.Join(testsuite.dbPath, "local"),
			Name:     "Folder2",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.dbPath, "local", "Folder2"),
			Name:     "File2",
			Size:     13,
			Hash:     "c28739a884e3742ea784f63dd52d9a4a90372235",
		},
		{
			IsFolder: true,
			Path:     filepath.Join(testsuite.dbPath, "local"),
			Name:     "Folder3",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.dbPath, "local", "Folder3"),
			Name:     "File3",
			Size:     13,
			Hash:     "995ea3c9a13945c2415a0881cc1bdac7a526b681",
		},
	}

	fsEntries, err := testsuite.store.GetLatestFileSystemEntries(testsuite.ctx, db.LocalFileSystem)
	testsuite.Require().NoError(err)
	testsuite.Require().Len(fsEntries, len(expected))

	sortedEntries := func(elements []db.FSEntry) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].EntryID < elements[j].EntryID }
	}

	sort.Slice(expected, sortedEntries(expected))
	sort.Slice(fsEntries, sortedEntries(fsEntries))

	// nolint: gocritic
	for i, e := range expected {
		actualE := fsEntries[i]
		testsuite.NotEmpty(actualE.DeviceID)
		testsuite.Greater(actualE.EntryID, uint64(0))
		testsuite.EqualValues(e.IsFolder, actualE.IsFolder)
		testsuite.EqualValues(e.Path, actualE.Path, "for: "+actualE.Name)
		testsuite.EqualValues(e.Name, actualE.Name)
		testsuite.Greaterf(actualE.ParentFolderID, uint64(0), "expected: %s - actual: %s", e.Name, fsEntries[i].Name)
		testsuite.WithinDuration(now, actualE.Created, 5*time.Minute)
		testsuite.WithinDuration(now, actualE.Modified, 5*time.Minute)
		if !e.IsFolder {
			testsuite.EqualValues(e.Size, actualE.Size)
		}
		testsuite.EqualValues(e.Hash, actualE.Hash)
	}
}

// nolint: gocritic
func (testsuite *IntegrationTestSuite) addNewFileSystemEntries(fse []db.FSEntry, fn func(e db.FSEntry) db.FSEntry) {
	if fn == nil {
		fn = func(e db.FSEntry) db.FSEntry { return e }
	}

	entriesCh, errCh := testsuite.store.AddNewFileSystemEntries(testsuite.ctx, db.WithEntriesChannelSize(0))

	err := func() error {
		for _, e := range fse {
			select {
			case err := <-errCh:
				close(entriesCh)
				return err
			case entriesCh <- fn(e):
			}
		}
		close(entriesCh)

		return <-errCh
	}()
	testsuite.Require().NoError(err)
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

// fsEntrySample1 counterpart to folderTreeSample1().
func fsEntrySample1(time1, time2, time3, time4, time5, time6, time7 time.Time) []db.FSEntry {
	return []db.FSEntry{
		{
			FSType:         db.PCloudFileSystem,
			EntryID:        0,
			IsFolder:       true,
			Path:           "/",
			Name:           "/",
			ParentFolderID: 0,
			Created:        time1,
			Modified:       time1,
		},
		{
			FSType:         db.PCloudFileSystem,
			EntryID:        20001,
			IsFolder:       true,
			Path:           "/",
			Name:           "Folder2",
			ParentFolderID: 0,
			Created:        time4,
			Modified:       time4,
		},
		{
			FSType:         db.PCloudFileSystem,
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
			FSType:         db.PCloudFileSystem,
			EntryID:        30001,
			IsFolder:       true,
			Path:           "/",
			Name:           "Folder3",
			ParentFolderID: 0,
			Created:        time6,
			Modified:       time6,
		},
		{
			FSType:         db.PCloudFileSystem,
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
