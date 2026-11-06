package analyzer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ArchiveLimits struct {
	MaxEntries      int
	MaxUncompressed int64
}

func ExtractZip(archivePath, target string, limits ArchiveLimits) (bool, error) {
	if limits.MaxEntries <= 0 {
		limits.MaxEntries = 20_000
	}
	if limits.MaxUncompressed <= 0 {
		limits.MaxUncompressed = 500 * 1024 * 1024
	}
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return false, fmt.Errorf("ZIP 格式不合法")
	}
	defer reader.Close()
	if len(reader.File) > limits.MaxEntries {
		return false, fmt.Errorf("源码归档文件数量超过限制")
	}

	var expanded int64
	hasJava := false
	for _, entry := range reader.File {
		name := filepath.Clean(strings.ReplaceAll(entry.Name, `\`, "/"))
		if name == "." || filepath.IsAbs(name) || name == ".." ||
			strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return false, fmt.Errorf("源码归档包含不安全路径")
		}
		if entry.Mode()&os.ModeSymlink != 0 {
			return false, fmt.Errorf("源码归档不允许符号链接")
		}
		destination := filepath.Join(target, filepath.FromSlash(name))
		relative, err := filepath.Rel(target, destination)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return false, fmt.Errorf("源码归档包含路径穿越")
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return false, fmt.Errorf("解压源码失败")
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return false, fmt.Errorf("解压源码失败")
		}
		source, err := entry.Open()
		if err != nil {
			return false, fmt.Errorf("源码条目无法读取")
		}
		file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			source.Close()
			return false, fmt.Errorf("解压源码失败")
		}
		written, copyErr := io.Copy(file, io.LimitReader(source, limits.MaxUncompressed-expanded+1))
		source.Close()
		file.Close()
		if copyErr != nil {
			return false, fmt.Errorf("解压源码失败")
		}
		expanded += written
		if expanded > limits.MaxUncompressed {
			return false, fmt.Errorf("源码解压后超过体积限制")
		}
		hasJava = hasJava || strings.HasSuffix(strings.ToLower(name), ".java")
	}
	return hasJava, nil
}
