package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir make sure the dir exists - checks that dir exists,
// if not creates it, including parent dirs.
func EnsureDir(outputDirPath string) {
	err := os.MkdirAll(outputDirPath, 0755)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		return
	}
}

// FindFiles traverses the directory tree starting at dir and returns files
// matching the wildcard pattern
func FindFiles(dir, wildcard string, inSubDirs bool) []string {
	var filePaths []string

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Directory '%s' does not exist\n", dir)
		os.Exit(1)
	}

	var err error
	if inSubDirs {
		err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				matched, err := filepath.Match(wildcard, d.Name())
				if err != nil {
					return err
				}
				if matched {
					filePaths = append(filePaths, path)
				}
			}
			return nil
		})
	} else {
		var matches []string
		matches, err = filepath.Glob(wildcard)

		filePaths = append(filePaths, matches...)
	}

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return filePaths
}
