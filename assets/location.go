// Package assets used to locate assets files
package assets

import (
	"path/filepath"
	"runtime"
)

var basePath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0) //nolint
	basePath = filepath.Dir(currentFile)
}

// Path return absolute path
func Path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}

	return filepath.Join(basePath, rel)
}
