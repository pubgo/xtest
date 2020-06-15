package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// Files returns all the Golang files for the given path. Ignores hidden files.
func Files(srcPath string) ([]Path, error) {
	srcPath, err := filepath.Abs(srcPath)
	if err != nil {
		return nil, fmt.Errorf("filepath.Abs: %v\n", err)
	}
	var fi os.FileInfo
	if fi, err = os.Stat(srcPath); err != nil {
		return nil, fmt.Errorf("os.Stat: %v\n", err)
	}
	if fi.IsDir() {
		return dirFiles(srcPath)
	}
	return file(srcPath)
}

func dirFiles(srcPath string) ([]Path, error) {
	ps, err := filepath.Glob(path.Join(srcPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("filepath.Glob: %v\n", err)
	}
	var srcPaths []Path
	for _, p := range ps {
		src := Path(p)
		if isHiddenFile(p) || src.IsTestPath() {
			continue
		}
		srcPaths = append(srcPaths, src)
	}
	return srcPaths, nil
}

func file(srcPath string) ([]Path, error) {
	src := Path(srcPath)
	if filepath.Ext(srcPath) != ".go" || isHiddenFile(srcPath) {
		return nil, fmt.Errorf("no Go source files found at %v", srcPath)
	}
	return []Path{src}, nil
}

func isHiddenFile(path string) bool {
	return []rune(filepath.Base(path))[0] == '.'
}
