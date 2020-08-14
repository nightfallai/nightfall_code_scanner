package gitdiff

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// GetDiff uses the command line to compute the diff
func GetDiff(workDir, baseRef, headRef string) (string, error) {
	err := os.Chdir(workDir)
	if err != nil {
		return "", err
	}
	err = exec.Command("git fetch", "origin", baseRef, "--depth=1").Run()
	if err != nil {
		return "", err
	}
	diffCmd := exec.Command("git diff", baseRef, headRef)
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
