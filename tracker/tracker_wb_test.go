package tracker

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHashFileData(t *testing.T) {
	const dbPath = "/tmp/data_test_hashFileData"

	err := os.MkdirAll(dbPath, 0700)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dbPath) }()

	fName := uuid.New().String()
	err = ioutil.WriteFile(filepath.Join(dbPath, fName), []byte("This is File000"), 0600)
	require.NoError(t, err)

	h, err := hashFileData(filepath.Join(dbPath, fName))
	require.NoError(t, err)
	assert.Equal(t, "01ce643e7c1ca98f6fb21e61b5d03f547813edae", h)
}
