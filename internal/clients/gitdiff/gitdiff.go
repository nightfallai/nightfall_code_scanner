package gitdiff

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

const unknownCommitHash = "0000000000000000000000000000000000000000"

// GitDiff client for getting diffs from the command line
type GitDiff struct {
	WorkDir string
	Base    string
	Head    string
}

// GetDiff uses the command line to compute the diff
func (gd *GitDiff) GetDiff() (string, error) {
	err := os.Chdir(gd.WorkDir)
	if err != nil {
		return "", err
	}
	err = exec.Command("git", "fetch", "origin", gd.Base, "--depth=1").Run()
	if err != nil {
		return "", err
	}

	var diffCmd *exec.Cmd
	if gd.Base == unknownCommitHash {
		diffCmd = exec.Command("git", "show", "--format=\"\"", gd.Head)
	} else {
		diffCmd = exec.Command("git", "diff", gd.Base, gd.Head)
	}
	reader, err := diffCmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	err = diffCmd.Start()
	if err != nil {
		return "", nil
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return "", err
	}
	err = diffCmd.Wait()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
