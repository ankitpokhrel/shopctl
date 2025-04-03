package registry

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
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
	var (
		located   = make(chan File)
		maxDepth  = 3
		baseDepth = strings.Count(filepath.Clean(dir), string(os.PathSeparator))
	)

	go func() {
		defer close(located)

		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				located <- File{Err: err}
				return nil
			}
			depth := strings.Count(filepath.Clean(path), string(os.PathSeparator)) - baseDepth
			if depth > maxDepth {
				return filepath.SkipDir
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

// ExtractZipToTemp extracts .tar.gz file to a temp location.
func ExtractZipToTemp(zipPath string, name string) (string, error) {
	zipFile, err := os.Open(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer func() { _ = zipFile.Close() }()

	gz, err := gzip.NewReader(zipFile)
	if err != nil {
		return "", err
	}
	defer func() { _ = gz.Close() }()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("shopctl-%s-*", name))
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	fileHandler := func(ctx context.Context, file archives.FileInfo) error {
		destPath := filepath.Join(tmpDir, file.NameInArchive)

		if file.IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directories for %s: %w", destPath, err)
		}

		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in zip: %w", file.NameInArchive, err)
		}
		defer func() { _ = src.Close() }()

		dest, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}
		defer func() { _ = dest.Close() }()

		if _, err := io.Copy(dest, src); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	}

	var zipFormat archives.Tar

	err = zipFormat.Extract(context.Background(), gz, fileHandler)
	if err != nil {
		return "", fmt.Errorf("failed to extract zip file: %w", err)
	}
	return tmpDir, nil
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
		return strings.HasSuffix(strings.TrimSuffix(info.Name(), ".tar.gz"), suffix)
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

		if !strings.HasSuffix(info.Name(), ".tar.gz") && !info.IsDir() {
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
