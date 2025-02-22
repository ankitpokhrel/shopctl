package registry

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
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

// GetAllInDir returns all files with given extension within a directory.
// It returns all folders if the ext is empty.
func GetAllInDir(dir, ext string) (<-chan File, error) {
	located := make(chan File)

	go func() {
		defer close(located)

		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				located <- File{Err: err}
				return nil
			}
			if ext != "" {
				if !info.IsDir() && strings.HasSuffix(info.Name(), ext) {
					located <- File{Path: path}
				}
			} else if info.IsDir() {
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

// LookForDir searches for a directory within a specified path.
func LookForDir(dir, in string) (string, error) {
	return lookForDir(in, func(info os.FileInfo) bool {
		return info.Name() == dir
	})
}

// LookForDirWithSuffix searches for a directory with a give suffix in a specified path.
func LookForDirWithSuffix(suffix, in string) (string, error) {
	return lookForDir(in, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), suffix)
	})
}

func lookForDir(in string, cmpFn func(os.FileInfo) bool) (string, error) {
	var (
		loc string

		maxDepth  = 5
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
		if cmpFn(info) {
			loc = path
			return ErrTargetFound // Stop walking.
		}
		return nil
	})
	if err != nil && err != ErrTargetFound {
		return "", err
	}
	if loc == "" {
		return "", ErrNoTargetFound
	}
	return loc, nil
}

// ReadFromTarGZ allow you to read a single file from the `.tar.gz` folder.
func ReadFromTarGZ(tarGzFilePath string, targetFileName string) ([]byte, error) {
	file, err := os.Open(tarGzFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer func() { _ = file.Close() }()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %w", err)
		}

		if header.Name == targetFileName {
			fileContent, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read file content: %w", err)
			}
			return fileContent, nil
		}
	}

	return nil, fmt.Errorf("file %s not found in the tar.gz archive", targetFileName)
}
