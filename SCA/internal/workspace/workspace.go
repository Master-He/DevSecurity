package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type Layout struct {
	Root      string
	ScanRoot  string
	InputRoot string
	Output    string
	Logs      string
	Ready     string
}

func New(root, scanID string) (Layout, error) {
	if strings.TrimSpace(root) == "" {
		return Layout{}, fmt.Errorf("SCA work root is required")
	}
	if !uuidPattern.MatchString(scanID) {
		return Layout{}, fmt.Errorf("scan_id must be a UUID")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve SCA work root: %w", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return Layout{}, fmt.Errorf("inspect SCA work root: %w", err)
	}
	if !info.IsDir() {
		return Layout{}, fmt.Errorf("SCA work root must be a directory")
	}
	scanRoot := filepath.Join(absoluteRoot, "scans", scanID)
	if !within(absoluteRoot, scanRoot) {
		return Layout{}, fmt.Errorf("scan workspace escapes SCA work root")
	}
	return Layout{
		Root:      absoluteRoot,
		ScanRoot:  scanRoot,
		InputRoot: filepath.Join(scanRoot, "input"),
		Output:    filepath.Join(scanRoot, "output"),
		Logs:      filepath.Join(scanRoot, "logs"),
		Ready:     filepath.Join(scanRoot, "READY"),
	}, nil
}

func (l Layout) ValidateInput(sourceKind string) (string, error) {
	if err := l.ensureExistingWithinRoot(l.ScanRoot); err != nil {
		return "", fmt.Errorf("scan workspace is missing or unsafe: %w", err)
	}
	if err := ensureRegular(l.Ready); err != nil {
		return "", fmt.Errorf("scan workspace READY marker is missing or invalid: %w", err)
	}
	if err := ensureDirectory(l.ScanRoot); err != nil {
		return "", err
	}
	switch sourceKind {
	case "directory":
		if err := ensureDirectory(l.InputRoot); err != nil {
			return "", fmt.Errorf("scan input directory is missing or invalid: %w", err)
		}
		if err := l.ensureExistingWithinRoot(l.InputRoot); err != nil {
			return "", fmt.Errorf("scan input directory is unsafe: %w", err)
		}
		return l.InputRoot, nil
	case "zip":
		input := filepath.Join(l.InputRoot, "source.zip")
		if err := ensureRegular(input); err != nil {
			return "", fmt.Errorf("scan source.zip is missing or invalid: %w", err)
		}
		if err := l.ensureExistingWithinRoot(input); err != nil {
			return "", fmt.Errorf("scan source.zip is unsafe: %w", err)
		}
		return input, nil
	default:
		return "", fmt.Errorf("unsupported source_kind %q", sourceKind)
	}
}

func (l Layout) EnsureOutputDirectories() error {
	for _, directory := range []string{l.Output, l.Logs} {
		if !within(l.Root, directory) {
			return fmt.Errorf("workspace path escapes SCA work root")
		}
		if err := os.MkdirAll(directory, 0o750); err != nil {
			return fmt.Errorf("create scan workspace directory: %w", err)
		}
		if err := ensureDirectory(directory); err != nil {
			return err
		}
		if err := l.ensureExistingWithinRoot(directory); err != nil {
			return fmt.Errorf("scan workspace directory is unsafe: %w", err)
		}
	}
	return nil
}

func (l Layout) ResultPath() string {
	return filepath.Join(l.Output, "result.json")
}

func (l Layout) ResultReference() string {
	relative, err := filepath.Rel(l.Root, l.ResultPath())
	if err != nil {
		return "output/result.json"
	}
	return filepath.ToSlash(relative)
}

func ensureRegular(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}
	return nil
}

func ensureDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("not a directory")
	}
	return nil
}

func within(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func (l Layout) ensureExistingWithinRoot(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !within(l.Root, absolute) {
		return fmt.Errorf("path escapes SCA work root")
	}
	relative, err := filepath.Rel(l.Root, absolute)
	if err != nil {
		return err
	}
	current := l.Root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path contains a symbolic link")
		}
	}
	return nil
}
