package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// This file contains a test that compiles and runs each program in testdata
// after generating the validation schema code for its type. The rule is that for testdata/x.go
// we run govader -type X and then compile and run the program. The resulting
// binary panics if the type.Validate() errors do not match expected errors.

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		os.Setenv("GOPACKAGESDEBUG", "true")
	}

	os.Exit(m.Run())
}

func TestEndToEnd(t *testing.T) {
	govader := govaderPath(t)
	// Read the testdata directory.
	fd, err := os.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	names, err := fd.Readdirnames(-1)
	if err != nil {
		t.Fatalf("Readdirnames: %s", err)
	}
	// Generate, compile, and run the test programs.
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			t.Errorf("%s is not a Go file", name)
			continue
		}
		t.Run(name, func(t *testing.T) {
			govaderCompileAndRun(t, t.TempDir(), govader, typeName(name), name)
		})
	}
}

// a type name for govader. use the last component of the file name with the .go
func typeName(fname string) string {
	// file names are known to be ascii and end .go
	base := path.Base(fname)
	return fmt.Sprintf("%c%s", base[0]+'A'-'a', base[1:len(base)-len(".go")])
}

var exe struct {
	path string
	err  error
	once sync.Once
}

func govaderPath(t *testing.T) string {
	return "govader" // TODO: return dynamic path from host.
	// exe.once.Do(func() {
	// 	exe.path, exe.err = os.Executable()
	// })
	// if exe.err != nil {
	// 	t.Fatal(exe.err)
	// }
	// return exe.path
}

// govaderCompileAndRun runs govader for the named file and compiles and
// runs the target binary in directory dir. That binary will panic if the String method is incorrect.
func govaderCompileAndRun(t *testing.T, dir, govader, typeName, fileName string) {
	t.Logf("run: %s %s\n", fileName, typeName)
	source := filepath.Join(dir, path.Base(fileName))
	err := copy(source, filepath.Join("testdata", fileName))
	if err != nil {
		t.Fatalf("copying file to temporary directory: %s", err)
	}
	schemaSource := filepath.Join(dir, typeName+"_schema.go")
	// Run govader in temporary directory.
	err = run(t, govader, "-type", typeName, "-output", schemaSource, source)
	if err != nil {
		t.Fatal(err)
	}
	// Run the binary in the temporary directory.
	err = run(t, "go", "run", schemaSource, source)
	if err != nil {
		t.Fatal(err)
	}
}

// copy copies the from file to the to file.
func copy(to, from string) error {
	toFd, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFd.Close()
	fromFd, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFd.Close()
	_, err = io.Copy(toFd, fromFd)
	return err
}

// run runs a single command and returns an error if it does not succeed.
// os/exec should have this function, to be honest.
func run(t testing.TB, name string, arg ...string) error {
	t.Helper()
	return runInDir(t, ".", name, arg...)
}

// runInDir runs a single command in directory dir and returns an error if
// it does not succeed.
func runInDir(t testing.TB, dir, name string, arg ...string) error {
	t.Helper()
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		t.Logf("%s", out)
	}
	if err != nil {
		return fmt.Errorf("%v: %v", cmd, err)
	}
	return nil
}
