package setup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// Give the parent a hidden console so the test also detects an unwanted
// inherited console when the test runner itself has none.
func TestScanCommandWithoutConsole(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	switch os.Getenv("ENOWX_TEST_SCAN_CONSOLE") {
	case "probe":
		console, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleWindow").Call()
		if console != 0 {
			fmt.Fprintln(os.Stderr, "background scan has a console")
			os.Exit(9)
		}
		fmt.Fprintln(os.Stdout, "stdout")
		fmt.Fprintln(os.Stderr, "stderr")
		os.Exit(7)
	case "parent":
		cmd := scanCommand(ctx, os.Args[0], "-test.run=^TestScanCommandWithoutConsole$")
		cmd.Env = append(os.Environ(), "ENOWX_TEST_SCAN_CONSOLE=probe")
		out, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 7 {
			t.Fatalf("exit = %v, output = %q", err, out)
		}
		if string(out) != "stdout\nstderr\n" {
			t.Fatalf("output = %q", out)
		}
		return
	}
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestScanCommandWithoutConsole$")
	cmd.Env = append(os.Environ(), "ENOWX_TEST_SCAN_CONSOLE=parent")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_CONSOLE, HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("console parent: %v\n%s", err, out)
	}
}
