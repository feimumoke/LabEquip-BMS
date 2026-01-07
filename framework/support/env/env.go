package env

import (
	"os"
	"path/filepath"
)

const WCRootDir = "WCROOT_DIR"

func GetProjectDir() string {
	rootDir := os.Getenv(WCRootDir)
	if len(rootDir) == 0 {
		wd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return wd
	}
	return rootDir
}
