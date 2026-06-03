package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the technocore binary once for all E2E tests
	tmpBin := filepath.Join(os.TempDir(), "technocore_e2e")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	build.Dir = repoRoot()
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n%s\n", err, out)
		os.Exit(1)
	}
	binaryPath = tmpBin
	code := m.Run()
	os.Remove(tmpBin)
	os.Exit(code)
}

func repoRoot() string {
	// Assume tests run from repo root or tests/e2e/
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "tests/e2e") {
		return filepath.Join(wd, "..", "..")
	}
	return wd
}

// e2eEnv holds per-test state: a temp HOME dir and helpers.
type e2eEnv struct {
	HomeDir string
	t       *testing.T
}

func newEnv(t *testing.T) *e2eEnv {
	home := t.TempDir()
	return &e2eEnv{HomeDir: home, t: t}
}

// run executes the technocore binary with args, returning stdout, stderr, and exit code.
func (e *e2eEnv) run(args ...string) (stdout, stderr string, exitCode int) {
	return e.runInDir(repoRoot(), args...)
}

// runInDir executes the binary in a specific working directory.
func (e *e2eEnv) runInDir(dir string, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+e.HomeDir)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// runWithInput executes the binary with piped stdin.
func (e *e2eEnv) runWithInput(stdin string, args ...string) (stdout, stderr string, exitCode int) {
	return e.runWithInputInDir(repoRoot(), stdin, args...)
}

// runWithInputInDir executes the binary with piped stdin in a specific directory.
func (e *e2eEnv) runWithInputInDir(dir, stdin string, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+e.HomeDir)
	cmd.Stdin = strings.NewReader(stdin)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// globalDBPath returns the path to the global DB in this env.
func (e *e2eEnv) globalDBPath() string {
	return filepath.Join(e.HomeDir, ".technocore", "global.db")
}
