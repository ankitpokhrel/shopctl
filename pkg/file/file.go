package file

import (
	"os"
	"path/filepath"
)

// Exists checks if a file exists at the specified location.
func Exists(loc string) bool {
	_, err := os.Stat(loc)
	return !os.IsNotExist(err)
}

// File represents a file.
type File struct {
	Path string
	Err  error
}

// FindFilesInDir searches for a file by name within a directory and its subdirectories.
func FindFilesInDir(dir, name string) (<-chan File, error) {
	located := make(chan File)

	go func() {
		defer close(located)

		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				located <- File{Err: err}
				return nil
			}
			if !info.IsDir() && info.Name() == name {
				located <- File{Path: path}
			}
			return nil
		})
	}()

	return located, nil
}

// ReadFileContents reads the contents of a file at a specified path.
func ReadFileContents(path string) ([]byte, error) {
	return os.ReadFile(path)
}
