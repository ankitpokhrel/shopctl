package registry

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

// GetLatestInDir returns the latest .tar.gz file within a directory.
func GetLatestInDir(dir string) (*File, string, error) {
	var (
		latest       *File
		latestTime   time.Time
		latestSuffix string
	)

	namePattern := regexp.MustCompile(`^.+_(\d{4}_\d{2}_\d{2}_\d{2}_\d{2}_\d{2})_(.+)\.tar\.gz$`)

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			latest = &File{Err: err}
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tar.gz") {
			matches := namePattern.FindStringSubmatch(info.Name())
			if matches == nil {
				return nil
			}
			datetime := matches[1]
			suffix := matches[2]

			parsedTime, err := time.Parse("2006_01_02_15_04_05", datetime)
			if err != nil {
				return fmt.Errorf("failed to parse datetime: %v", err)
			}

			if parsedTime.After(latestTime) {
				latest = &File{Path: path}
				latestTime = parsedTime
				latestSuffix = suffix
			}
		}
		return nil
	})
	if latest != nil {
		return latest, latestSuffix, nil
	}
	return nil, "", ErrNoTargetFound
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
