package filesystem_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/seborama/pcloud/tracker/db"
	"github.com/seborama/pcloud/tracker/filesystem"
)

type LocalIntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	localTestPath string

	localFS *filesystem.Local
}

func TestLocalIntegrationSuite(t *testing.T) {
	suite.Run(t, new(LocalIntegrationTestSuite))
}

func (testsuite *LocalIntegrationTestSuite) SetupSuite() {
	testsuite.ctx = context.Background()

	localTestPath, err := ioutil.TempDir("", "go_pCloud_test")
	testsuite.Require().NoError(err)

	testsuite.localTestPath = localTestPath
}

func (testsuite *LocalIntegrationTestSuite) TearDownSuite() {
	_ = os.RemoveAll(testsuite.localTestPath)
}

func (testsuite *LocalIntegrationTestSuite) BeforeTest(suiteName, testName string) {
	localFS := filesystem.NewLocal()
	testsuite.localFS = localFS
}

func (testsuite *LocalIntegrationTestSuite) AfterTest(suiteName, testName string) {
	_ = os.RemoveAll(testsuite.localTestPath)
}

func (testsuite *LocalIntegrationTestSuite) TestLocal_Walk() {
	now := time.Now()

	err := os.MkdirAll(filepath.Join(testsuite.localTestPath, "local"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.localTestPath, "local", "File000"), []byte("This is File000"), 0600)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(filepath.Join(testsuite.localTestPath, "local", "Folder1"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.localTestPath, "local", "Folder1", "File1"), []byte("This is File1"), 0600)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(filepath.Join(testsuite.localTestPath, "local", "Folder2"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.localTestPath, "local", "Folder2", "File2"), []byte("This is File2"), 0600)
	testsuite.Require().NoError(err)

	err = os.MkdirAll(filepath.Join(testsuite.localTestPath, "local", "Folder3"), 0700)
	testsuite.Require().NoError(err)
	err = ioutil.WriteFile(filepath.Join(testsuite.localTestPath, "local", "Folder3", "File3"), []byte("This is File3"), 0600)
	testsuite.Require().NoError(err)

	fsEntriesCh := make(chan db.FSEntry)
	errCh := make(chan error)
	fsEntries := []db.FSEntry{}

	go func() {
		for fse := range fsEntriesCh {
			fsEntries = append(fsEntries, fse)
		}

		errCh <- nil
	}()

	err = testsuite.localFS.Walk(testsuite.ctx, "local_fs", filepath.Join(testsuite.localTestPath, "local"), fsEntriesCh, errCh)
	testsuite.Require().NoError(err)

	expected := []db.FSEntry{
		{
			IsFolder: true,
			Path:     testsuite.localTestPath,
			Name:     "local",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.localTestPath, "local"),
			Name:     "File000",
			Size:     15,
			Hash:     "01ce643e7c1ca98f6fb21e61b5d03f547813edae",
		},
		{
			IsFolder: true,
			Path:     filepath.Join(testsuite.localTestPath, "local"),
			Name:     "Folder1",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.localTestPath, "local", "Folder1"),
			Name:     "File1",
			Size:     13,
			Hash:     "e8dfb879ddc708ea337a00e9b5580b498193bd2d",
		},
		{
			IsFolder: true,
			Path:     filepath.Join(testsuite.localTestPath, "local"),
			Name:     "Folder2",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.localTestPath, "local", "Folder2"),
			Name:     "File2",
			Size:     13,
			Hash:     "c28739a884e3742ea784f63dd52d9a4a90372235",
		},
		{
			IsFolder: true,
			Path:     filepath.Join(testsuite.localTestPath, "local"),
			Name:     "Folder3",
			Hash:     "",
		},
		{
			IsFolder: false,
			Path:     filepath.Join(testsuite.localTestPath, "local", "Folder3"),
			Name:     "File3",
			Size:     13,
			Hash:     "995ea3c9a13945c2415a0881cc1bdac7a526b681",
		},
	}

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
