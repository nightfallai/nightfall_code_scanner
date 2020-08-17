package gitdiff

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const unknownCommitHash = "0000000000000000000000000000000000000000"

// GitDiff client for getting diffs from the command line
type GitDiff struct {
	WorkDir    string
	BaseBranch string
	BaseSHA    string
	Head       string
}

// GetDiff uses the command line to compute the diff
func (gd *GitDiff) GetDiff() (string, error) {
	err := os.Chdir(gd.WorkDir)
	if err != nil {
		return "", err
	}
	fmt.Println("BaseBranch:", gd.BaseBranch)
	fmt.Println("BaseSHA:", gd.BaseSHA)
	fmt.Println("Head:", gd.Head)

	var diffCmd *exec.Cmd
	if gd.BaseBranch != "" {
		err = exec.Command("git", "fetch", "origin", gd.BaseBranch, "--depth=1").Run()
		if err != nil {
			fmt.Println("fetch base", gd.BaseBranch)
			return "", err
		}
		diffCmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s", gd.BaseBranch), gd.Head)
	}
	if gd.BaseSHA == "" || gd.BaseSHA == unknownCommitHash {
		err = exec.Command("git", "fetch", "origin", gd.Head, "--depth=2").Run()
		if err != nil {
			return "", err
		}
		diffCmd = exec.Command("git", "diff", "HEAD^", "HEAD")
	} else {
		err = exec.Command("git", "fetch", "origin", gd.BaseSHA, "--depth=1").Run()
		if err != nil {
			fmt.Println("fetch base", gd.BaseSHA)
			return "", err
		}
		diffCmd = exec.Command("git", "diff", gd.BaseSHA, gd.Head)
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
