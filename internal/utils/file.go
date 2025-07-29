package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RemoveBinaries attempts to remove the given list of binary names under the specified dir.
// It returns a combined error if any removal fails.
func RemoveBinaries(binDir string, names []string) error {
	var errList []string
	for _, name := range names {
		path := filepath.Join(binDir, name)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errList = append(errList, fmt.Sprintf("remove %s: %v", path, err))
		}
	}

	if len(errList) > 0 {
		return fmt.Errorf("errors removing binaries: %s", strings.Join(errList, "; "))
	}
	return nil
}
