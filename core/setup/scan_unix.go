//go:build !windows

package setup

import (
	"context"
	"os/exec"
)

func scanCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
