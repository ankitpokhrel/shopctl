package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFilesInDir(t *testing.T) {
	path := "./testdata/.tmp/"

	testFile := "testfile.txt"
	testFiles := []string{
		filepath.Join(path, testFile),
		filepath.Join(path, "subdir1", testFile),
		filepath.Join(path, "subdir2", "nested", testFile),
	}

	for _, filePath := range testFiles {
		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0o755)
		assert.NoError(t, err, "Failed to create directory %s", dir)

		err = os.WriteFile(filePath, []byte("test content"), 0o644)
		assert.NoError(t, err, "Failed to create test file %s", filePath)
	}

	locatedFiles, err := FindFilesInDir(path, testFile)
	assert.NoError(t, err)

	results := make([]string, 0)
	for file := range locatedFiles {
		assert.NoError(t, file.Err)
		results = append(results, file.Path)
	}

	expectedResults := make(map[string]bool)
	for _, path := range testFiles {
		expectedResults[path] = true
	}
	assert.Equal(t, len(expectedResults), len(results))

	for _, result := range results {
		assert.True(t, expectedResults[result])
		delete(expectedResults, result)
	}
	assert.Empty(t, expectedResults)

	// Clean up.
	assert.NoError(t, os.RemoveAll(path))
}

func TestLookForDir(t *testing.T) {
	path := "./testdata/bkp/"

	loc, err := LookForDir("eg", path)
	assert.NoError(t, err)
	assert.Equal(t, "testdata/bkp/2025/01/eg", loc)
}

func TestLookForDirWithSuffix(t *testing.T) {
	path := "./testdata/bkp/"

	loc, err := LookForDirWithSuffix("1", path)
	assert.NoError(t, err)
	assert.Equal(t, "testdata/bkp/2025/01", loc)
}
