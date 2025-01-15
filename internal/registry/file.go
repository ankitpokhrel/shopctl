package registry

import (
	"os"
	"path/filepath"
	"strings"
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

// LookForDir searches for a directory within a specified path.
func LookForDir(dir, in string) (string, error) {
	var (
		loc string

		maxDepth  = 4
		baseDepth = strings.Count(filepath.Clean(in), string(os.PathSeparator))
	)

	err := filepath.Walk(in, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		depth := strings.Count(filepath.Clean(path), string(os.PathSeparator)) - baseDepth

		if !info.IsDir() {
			return nil
		}
		if depth > maxDepth {
			return filepath.SkipDir
		}
		if info.Name() == dir {
			loc = path
			return ErrTargetFound // Stop walking.
		}
		return nil
	})
	if err != nil && err != ErrTargetFound {
		return "", err
	}
	return loc, nil
}

// ReadFileContents reads the contents of a file at a specified path.
func ReadFileContents(path string) ([]byte, error) {
	return os.ReadFile(path)
}
