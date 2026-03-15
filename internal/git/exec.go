package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

const defaultTimeout = 30 * time.Second

// RunGit executes a git command in the repo root and returns stdout.
func (h *HistoryAnalyzer) RunGit(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = h.repoRoot

	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %v failed: %w (stderr: %s)", args, err, errOut.String())
	}

	return out.String(), nil
}
