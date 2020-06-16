// Package gotests contains the core logic for generating table-driven tests.
package internal

import (
	"fmt"
	"go/importer"
	"go/types"
	"os"
	"path"
	"regexp"
	"sort"
	"sync"
)

// Options provides custom filters and parameters for generating tests.
type Options struct {
	Only           *regexp.Regexp         // Includes only functions that match.
	Exclude        *regexp.Regexp         // Excludes functions that match.
	Exported       bool                   // Include only exported methods
	PrintInputs    bool                   // Print function parameters in error messages
	Subtests       bool                   // Print tests using Go 1.7 subtests
	Parallel       bool                   // Print tests that runs the subtests in parallel.
	Importer       func() types.Importer  // A custom importer.
	Template       string                 // Name of custom template set
	TemplateDir    string                 // Path to custom template set
	TemplateParams map[string]interface{} // Custom external parameters
}

// A GeneratedTest contains information about a test file with generated tests.
type GeneratedTest struct {
	Path      string      // The test file's absolute path.
	Functions []*Function // The functions with new test methods.
}

// GenerateTests generates table-driven tests for the function and method
// signatures defined in the target source path file(s). The source path
// parameter can be either a Go source file or directory containing Go files.
func GenerateTests(srcPath string, opt *Options) ([]*GeneratedTest, error) {
	if opt == nil {
		opt = &Options{}
	}
	srcFiles, err := Files(srcPath)
	if err != nil {
		return nil, fmt.Errorf("Files: %v", err)
	}
	files, err := Files(path.Dir(srcPath))
	if err != nil {
		return nil, fmt.Errorf("Files: %v", err)
	}
	if opt.Importer == nil || opt.Importer() == nil {
		opt.Importer = importer.Default
	}
	return parallelize(srcFiles, files, opt)
}

// result stores a generateTest result.
type result struct {
	gt  *GeneratedTest
	err error
}

// parallelize generates tests for the given source files concurrently.
func parallelize(srcFiles, files []Path, opt *Options) ([]*GeneratedTest, error) {
	var wg sync.WaitGroup
	rs := make(chan *result, len(srcFiles))
	for _, src := range srcFiles {
		wg.Add(1)
		// Worker
		go func(src Path) {
			defer wg.Done()
			r := &result{}
			r.gt, r.err = generateTest(src, files, opt)
			rs <- r
		}(src)
	}
	// Closer.
	go func() {
		wg.Wait()
		close(rs)
	}()
	return readResults(rs)
}

// readResults reads the result channel.
func readResults(rs <-chan *result) ([]*GeneratedTest, error) {
	var gts []*GeneratedTest
	for r := range rs {
		if r.err != nil {
			return nil, r.err
		}
		if r.gt != nil {
			gts = append(gts, r.gt)
		}
	}
	return gts, nil
}

func generateTest(src Path, files []Path, opt *Options) (*GeneratedTest, error) {
	p := &Parser{Importer: opt.Importer()}
	sr, err := p.Parse(string(src), files)
	if err != nil {
		return nil, fmt.Errorf("Parser.Parse source file: %v", err)
	}
	h := sr.Header
	h.Code = nil // Code is only needed from parsed test files.
	testPath := src.TestPath()
	h, tf, err := parseTestFile(p, testPath, h)
	if err != nil {
		return nil, err
	}
	funcs := testableFuncs(sr.Funcs, opt.Only, opt.Exclude, opt.Exported, tf)
	if len(funcs) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("output.Process: %v", err)
	}
	return &GeneratedTest{
		Path:      testPath,
		Functions: funcs,
	}, nil
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func parseTestFile(p *Parser, testPath string, h *Header) (*Header, []string, error) {
	if !IsFileExist(testPath) {
		return h, nil, nil
	}
	tr, err := p.Parse(testPath, nil)
	if err != nil {
		if err == ErrEmptyFile {
			// Overwrite empty test files.
			return h, nil, nil
		}
		return nil, nil, fmt.Errorf("Parser.Parse test file: %v", err)
	}
	var testFuncs []string
	for _, fun := range tr.Funcs {
		testFuncs = append(testFuncs, fun.Name)
	}
	tr.Header.Imports = append(tr.Header.Imports, h.Imports...)
	h = tr.Header
	return h, testFuncs, nil
}

func testableFuncs(funcs []*Function, only, excl *regexp.Regexp, exp bool, testFuncs []string) []*Function {
	sort.Strings(testFuncs)
	var fs []*Function
	for _, f := range funcs {
		if isTestFunction(f, testFuncs) || isExcluded(f, excl) || isUnexported(f, exp) || !isIncluded(f, only) || isInvalid(f) {
			continue
		}
		fs = append(fs, f)
	}
	return fs
}

func isInvalid(f *Function) bool {
	if f.Name == "init" && f.IsNaked() {
		return true
	}
	return false
}

func isTestFunction(f *Function, testFuncs []string) bool {
	return len(testFuncs) > 0 && contains(testFuncs, f.TestName())
}

func isExcluded(f *Function, excl *regexp.Regexp) bool {
	return excl != nil && (excl.MatchString(f.Name) || excl.MatchString(f.FullName()))
}

func isUnexported(f *Function, exp bool) bool {
	return exp && !f.IsExported
}

func isIncluded(f *Function, only *regexp.Regexp) bool {
	return only == nil || only.MatchString(f.Name) || only.MatchString(f.FullName())
}

func contains(ss []string, s string) bool {
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		return true
	}
	return false
}
