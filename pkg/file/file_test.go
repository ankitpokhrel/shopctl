package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFilesInDir(t *testing.T) {
	path := "./testdata/.tmp/"

	// Create a nested directory structure and test files
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
	assert.NoError(t, err, "Unexpected error during function call")

	results := make([]string, 0)
	for file := range locatedFiles {
		assert.NoError(t, file.Err)
		results = append(results, file.Path)
	}

	expectedResults := make(map[string]bool)
	for _, path := range testFiles {
		expectedResults[path] = true
	}
	assert.Equal(t, len(expectedResults), len(results), "Mismatch in number of results")

	for _, result := range results {
		assert.True(t, expectedResults[result], "Unexpected file path in results: %s", result)
		delete(expectedResults, result)
	}
	assert.Empty(t, expectedResults, "Missing expected results: %v", expectedResults)

	// Clean up.
	assert.NoError(t, os.RemoveAll(path))
}
