package source

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultMaxArchiveBytes     int64 = 20 << 30
	defaultMaxUncompressedSize int64 = 20 << 30
	defaultMaxFiles                  = 200000
)

type Limits struct {
	MaxArchiveBytes     int64
	MaxUncompressedSize int64
	MaxFiles            int
}

type Prepared struct {
	Root      string
	temporary bool
}

func (p *Prepared) Close() error {
	if p == nil || !p.temporary || p.Root == "" {
		return nil
	}
	return os.RemoveAll(p.Root)
}

func DefaultLimits() Limits {
	return Limits{
		MaxArchiveBytes:     envInt64("SCA_MAX_ARCHIVE_BYTES", defaultMaxArchiveBytes),
		MaxUncompressedSize: envInt64("SCA_MAX_UNCOMPRESSED_BYTES", defaultMaxUncompressedSize),
		MaxFiles:            int(envInt64("SCA_MAX_ARCHIVE_FILES", int64(defaultMaxFiles))),
	}
}

func Prepare(input string, limits Limits) (*Prepared, error) {
	return PrepareIn(input, "", limits)
}

func PrepareIn(input, temporaryParent string, limits Limits) (*Prepared, error) {
	if strings.TrimSpace(input) == "" {
		return nil, errors.New("input path is required")
	}
	absolute, err := filepath.Abs(input)
	if err != nil {
		return nil, fmt.Errorf("resolve input path: %w", err)
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return nil, fmt.Errorf("inspect input: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("input path must not be a symbolic link")
	}
	if info.IsDir() {
		return &Prepared{Root: absolute}, nil
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("input must be a directory or ZIP file")
	}
	if limits.MaxArchiveBytes <= 0 {
		limits.MaxArchiveBytes = defaultMaxArchiveBytes
	}
	if limits.MaxUncompressedSize <= 0 {
		limits.MaxUncompressedSize = defaultMaxUncompressedSize
	}
	if limits.MaxFiles <= 0 {
		limits.MaxFiles = defaultMaxFiles
	}
	if info.Size() > limits.MaxArchiveBytes {
		return nil, fmt.Errorf("input archive exceeds %d bytes", limits.MaxArchiveBytes)
	}

	root, err := os.MkdirTemp(temporaryParent, ".devsecurity-sca-source-*")
	if err != nil {
		return nil, fmt.Errorf("create source workspace: %w", err)
	}
	prepared := &Prepared{Root: root, temporary: true}
	if err := extractZip(absolute, root, limits); err != nil {
		prepared.Close()
		return nil, err
	}
	return prepared, nil
}

func extractZip(archivePath, target string, limits Limits) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open ZIP input: %w", err)
	}
	defer reader.Close()
	if len(reader.File) > limits.MaxFiles {
		return fmt.Errorf("ZIP contains more than %d entries", limits.MaxFiles)
	}

	var total int64
	seen := make(map[string]bool, len(reader.File))
	for _, entry := range reader.File {
		name := strings.ReplaceAll(entry.Name, `\`, "/")
		clean := path.Clean(name)
		if clean == "." || clean == "" || path.IsAbs(clean) ||
			clean == ".." || strings.HasPrefix(clean, "../") ||
			filepath.VolumeName(filepath.FromSlash(clean)) != "" {
			return fmt.Errorf("ZIP contains an unsafe path")
		}
		if seen[clean] {
			return fmt.Errorf("ZIP contains a duplicate path")
		}
		seen[clean] = true
		if entry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("ZIP contains a symbolic link")
		}

		destination := filepath.Join(target, filepath.FromSlash(clean))
		relative, err := filepath.Rel(target, destination)
		if err != nil || relative == ".." ||
			strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("ZIP contains a path traversal")
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return fmt.Errorf("extract ZIP directory: %w", err)
			}
			continue
		}
		if !entry.Mode().IsRegular() {
			return fmt.Errorf("ZIP contains an unsupported file type")
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return fmt.Errorf("extract ZIP directory: %w", err)
		}
		source, err := entry.Open()
		if err != nil {
			return fmt.Errorf("read ZIP entry: %w", err)
		}
		destinationFile, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			source.Close()
			return fmt.Errorf("create extracted file: %w", err)
		}
		remaining := limits.MaxUncompressedSize - total + 1
		if remaining <= 0 {
			source.Close()
			destinationFile.Close()
			return fmt.Errorf("ZIP expands beyond %d bytes", limits.MaxUncompressedSize)
		}
		written, copyErr := io.Copy(destinationFile, io.LimitReader(source, remaining))
		source.Close()
		closeErr := destinationFile.Close()
		total += written
		if copyErr != nil {
			return fmt.Errorf("extract ZIP entry: %w", copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close extracted file: %w", closeErr)
		}
		if total > limits.MaxUncompressedSize {
			return fmt.Errorf("ZIP expands beyond %d bytes", limits.MaxUncompressedSize)
		}
	}
	return nil
}

func envInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
