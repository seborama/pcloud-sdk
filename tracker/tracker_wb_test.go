package tracker

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHashFileData(t *testing.T) {
	err := os.MkdirAll("data_test_hashFileData", 0x700)
	require.NoError(t, err)

	fName := uuid.New().String()
	err = ioutil.WriteFile("data_test_hashFileData/"+fName, []byte("This is File000"), 0x700)
	require.NoError(t, err)
	defer func() {
		os.RemoveAll("data_test_hashFileData/" + fName)
	}()

	h, err := hashFileData("data_test_hashFileData/" + fName)
	require.NoError(t, err)
	assert.Equal(t, "01ce643e7c1ca98f6fb21e61b5d03f547813edae", h)
}
