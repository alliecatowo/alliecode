package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func readJSONState(path, label string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", label, err)
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode %s: %w", label, err)
	}
	return nil
}

func writeJSONState(path, label string, state any) error {
	if err := os.MkdirAll(filepath.Dir(path), stateDirPerm); err != nil {
		return fmt.Errorf("create %s directory: %w", label, err)
	}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", label, err)
	}
	payload = append(payload, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, stateFilePerm); err != nil {
		return fmt.Errorf("write %s tmp file: %w", label, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace %s file: %w", label, err)
	}
	return nil
}
