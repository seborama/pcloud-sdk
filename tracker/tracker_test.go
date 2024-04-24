package tracker_test

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/seborama/pcloud-sdk/tracker"
	"github.com/seborama/pcloud-sdk/tracker/db"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	dbPath string
	store  *db.SQLite3

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
	testsuite.makeDB()

	track, err := tracker.NewTracker(testsuite.ctx, zap.NewNop(), testsuite.store, nil /*filesystem.NewLocal()*/, "some_fs")
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

func (testsuite *IntegrationTestSuite) TestFindMutations_FilesDeleted() {
	time1 := time.Now().Add(-24 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.MarkFileSystemAsChanged(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	expected := db.FSMutations{
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
						EntryID:        0,
						IsFolder:       true,
						Path:           "/",
						Name:           "/",
						ParentFolderID: 0,
						Created:        time1,
						Modified:       time1,
					},
				},
			},
		},
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
						EntryID:        20001,
						IsFolder:       true,
						Path:           "/",
						Name:           "Folder2",
						ParentFolderID: 0,
						Created:        time4,
						Modified:       time4,
					},
				},
			},
		},
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
			},
		},
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
						EntryID:        30001,
						IsFolder:       true,
						Path:           "/",
						Name:           "Folder3",
						ParentFolderID: 0,
						Created:        time6,
						Modified:       time6,
					},
				},
			},
		},
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
			},
		},
	}

	fsMutations, err := testsuite.tracker.ListMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements db.FSMutations) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].Details[0].EntryID < elements[j].Details[0].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindMutations_FilesCreated() {
	time1 := time.Now().Add(-24 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.RotateFileSystemVersions(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	err = testsuite.store.MarkFileSystemAsChanged(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	expected := db.FSMutations{
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
						EntryID:        0,
						IsFolder:       true,
						Path:           "/",
						Name:           "/",
						ParentFolderID: 0,
						Created:        time1,
						Modified:       time1,
					},
				},
			},
		},
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
						EntryID:        20001,
						IsFolder:       true,
						Path:           "/",
						Name:           "Folder2",
						ParentFolderID: 0,
						Created:        time4,
						Modified:       time4,
					},
				},
			},
		},
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
			},
		},
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
						EntryID:        30001,
						IsFolder:       true,
						Path:           "/",
						Name:           "Folder3",
						ParentFolderID: 0,
						Created:        time6,
						Modified:       time6,
					},
				},
			},
		},
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
			},
		},
	}

	fsMutations, err := testsuite.tracker.ListMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements db.FSMutations) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].Details[0].EntryID < elements[j].Details[0].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindMutations_FileModified() {
	time1 := time.Now().Add(-24 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.RotateFileSystemVersions(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	fse1 = fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, func(e db.FSEntry) db.FSEntry {
		if e.EntryID == 20002 {
			e.Hash = "1234565432100020002"
		}
		return e
	})

	err = testsuite.store.MarkFileSystemAsChanged(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	expected := db.FSMutations{
		{
			Type: db.MutationTypeModified,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
			},
		},
	}

	fsMutations, err := testsuite.tracker.ListMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements db.FSMutations) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].Details[0].EntryID < elements[j].Details[0].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindMutations_FileMoved() {
	time1 := time.Now().Add(-24 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.RotateFileSystemVersions(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	fse1 = fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, func(e db.FSEntry) db.FSEntry {
		if e.EntryID == 20002 {
			// move File2 to Folder3
			e.Path = "/Folder3"
			e.ParentFolderID = 30001
		}
		return e
	})

	err = testsuite.store.MarkFileSystemAsChanged(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	expected := db.FSMutations{
		{
			Type: db.MutationTypeMoved,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "some_fs",
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
			},
		},
	}

	fsMutations, err := testsuite.tracker.ListMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements db.FSMutations) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].Details[0].EntryID < elements[j].Details[0].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
	}
}

func (testsuite *IntegrationTestSuite) TestFindMutations_NoChanges() {
	time1 := time.Now().Add(-24 * time.Hour)
	time4 := time.Now().Add(-21 * time.Hour)
	time5 := time.Now().Add(-20 * time.Hour)
	time6 := time.Now().Add(-19 * time.Hour)
	time7 := time.Now().Add(-18 * time.Hour)

	fse1 := fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err := testsuite.store.RotateFileSystemVersions(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	fse1 = fsEntrySample1(time1, time4, time5, time6, time7)

	testsuite.addNewFileSystemEntries(fse1, nil)

	err = testsuite.store.MarkFileSystemAsChanged(testsuite.ctx, "some_fs")
	testsuite.Require().NoError(err)

	expected := db.FSMutations{}

	fsMutations, err := testsuite.tracker.ListMutations(testsuite.ctx)
	testsuite.Require().NoError(err)

	sortedMutations := func(elements db.FSMutations) func(i, j int) bool {
		return func(i, j int) bool { return elements[i].Details[0].EntryID < elements[j].Details[0].EntryID }
	}
	sort.Slice(expected, sortedMutations(expected))
	sort.Slice(fsMutations, sortedMutations(fsMutations))
	if d := cmp.Diff(expected, fsMutations, cmpopts.IgnoreUnexported()); d != "" {
		testsuite.Fail(d)
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

// fsEntrySample1 counterpart to folderTreeSample1().
func fsEntrySample1(time1, time4, time5, time6, time7 time.Time) []db.FSEntry {
	return []db.FSEntry{
		{
			FSName:         "some_fs",
			EntryID:        0,
			IsFolder:       true,
			Path:           "/",
			Name:           "/",
			ParentFolderID: 0,
			Created:        time1,
			Modified:       time1,
		},
		{
			FSName:         "some_fs",
			EntryID:        20001,
			IsFolder:       true,
			Path:           "/",
			Name:           "Folder2",
			ParentFolderID: 0,
			Created:        time4,
			Modified:       time4,
		},
		{
			FSName:         "some_fs",
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
			FSName:         "some_fs",
			EntryID:        30001,
			IsFolder:       true,
			Path:           "/",
			Name:           "Folder3",
			ParentFolderID: 0,
			Created:        time6,
			Modified:       time6,
		},
		{
			FSName:         "some_fs",
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
