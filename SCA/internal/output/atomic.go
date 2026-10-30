package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteJSON(path string, value any) error {
	if path == "" {
		return fmt.Errorf("output path is required")
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	file, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary result: %w", err)
	}
	temporary := file.Name()
	defer os.Remove(temporary)

	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("set result permissions: %w", err)
	}
	if err := json.NewEncoder(file).Encode(value); err != nil {
		file.Close()
		return fmt.Errorf("encode result: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync result: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close result: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("publish result: %w", err)
	}
	return nil
}
