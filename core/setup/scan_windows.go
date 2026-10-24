package setup

import (
	"context"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func scanCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	// Tool probes run in the background, outside interactive sessions.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW}
	return cmd
}
