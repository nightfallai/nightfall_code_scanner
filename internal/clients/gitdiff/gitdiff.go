package gitdiff

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	githublogger "github.com/nightfallai/nightfall_code_scanner/internal/clients/logger/github_logger"
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
	logger := githublogger.NewGithubLogger(log.New(os.Stdout, "", 0))
	logger.Info("Getting Diff...")
	err := os.Chdir(gd.WorkDir)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("GitDiff: %+v", *gd))
	var diffCmd *exec.Cmd
	switch {
	case gd.BaseBranch != "":
		// PR event so get diff between base branch and current commit SHA
		logger.Info("Getting Diff between Base Branch and Current Commit SHA")
		//err = exec.Command("git", "fetch", "origin", gd.BaseBranch, "--depth=1").Run()
		fetchCmd := exec.Command(
			"GIT_TRACE=true",
			"GIT_CURL_VERBOSE=true",
			"GIT_SSH_COMMAND=\"ssh -vvv\"\t",
			"GIT_TRACE_SHALLOW=true",
			"git", "-c", "http.sslVerify=false", "fetch", "origin", gd.BaseBranch, "--depth=1")
		reader, err := fetchCmd.StdoutPipe()
		if err != nil {
			logger.Error(fmt.Sprintf("Error piping git fetch cmd: %s", err.Error()))
			return "", err
		}
		defer reader.Close()
		err = fetchCmd.Start()
		if err != nil {
			logger.Error(fmt.Sprintf("Error starting git diff cmd: %s", err.Error()))
			return "", nil
		}
		buf := new(strings.Builder)
		_, err = io.Copy(buf, reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Error copying git diff output: %s", err.Error()))
			return "", err
		}
		err = fetchCmd.Wait()
		if err != nil {
			logger.Error(fmt.Sprintf("Error waiting for git diff cmd to exit: %s", err.Error()))
			return "", err
		}
		logger.Info(fmt.Sprintf("Fetch Origin Logs: %s", buf.String()))
		/*err = exec.Command("git", "-c", "http.sslVerify=false", "fetch", "origin", gd.BaseBranch, "--depth=1").Run()
		if err != nil {
			logger.Error(fmt.Sprintf("Error getting diff between Base Branch and Current Commit SHA: %s", err.Error()))
			return "", err
		}*/
		diffCmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s", gd.BaseBranch))
	case gd.BaseSHA == "" || gd.BaseSHA == unknownCommitHash:
		// PUSH event for new branch so use git show to get the diff of the most recent commit
		logger.Info("Getting Diff for new branch push event")
		//err = exec.Command("git", "fetch", "origin", gd.Head, "--depth=2").Run()
		err = exec.Command("git", "-c", "http.sslVerify=false", "fetch", "origin", gd.Head, "--depth=2").Run()
		if err != nil {
			logger.Error(fmt.Sprintf("Error getting diff in new branch push event: %s", err.Error()))
			return "", err
		}
		diffCmd = exec.Command("git", "show", gd.Head, "--format=")
	default:
		// PUSH event where last commit action ran on exists
		// use current commit SHA and previous action run commit SHA for diff
		logger.Info("Getting Diff for new push event")
		//err = exec.Command("git", "fetch", "origin", gd.BaseSHA, "--depth=1").Run()
		fetchCmd := exec.Command(
			"GIT_TRACE=true",
			"GIT_CURL_VERBOSE=true",
			"GIT_SSH_COMMAND=\"ssh -vvv\"\t",
			"GIT_TRACE_SHALLOW=true",
			"git", "-c", "http.sslVerify=false", "fetch", "origin", gd.BaseSHA, "--depth=1")
		reader, err := fetchCmd.StdoutPipe()
		if err != nil {
			logger.Error(fmt.Sprintf("Error piping git fetch cmd: %s", err.Error()))
			return "", err
		}
		defer reader.Close()
		err = fetchCmd.Start()
		if err != nil {
			logger.Error(fmt.Sprintf("Error starting git diff cmd: %s", err.Error()))
			return "", nil
		}
		buf := new(strings.Builder)
		_, err = io.Copy(buf, reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Error copying git diff output: %s", err.Error()))
			return "", err
		}
		err = fetchCmd.Wait()
		if err != nil {
			logger.Error(fmt.Sprintf("Error waiting for git diff cmd to exit: %s", err.Error()))
			return "", err
		}
		logger.Info(fmt.Sprintf("Fetch Origin Logs: %s", buf.String()))

		/*if err != nil {
			logger.Error(fmt.Sprintf("Error fetching diff in push event: %s", err.Error()))
			return "", err
		}*/
		diffCmd = exec.Command("git", "diff", gd.BaseSHA, gd.Head)
	}

	reader, err := diffCmd.StdoutPipe()
	if err != nil {
		logger.Error(fmt.Sprintf("Error piping git diff cmd: %s", err.Error()))
		return "", err
	}
	defer reader.Close()

	err = diffCmd.Start()
	if err != nil {
		logger.Error(fmt.Sprintf("Error starting git diff cmd: %s", err.Error()))
		return "", nil
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		logger.Error(fmt.Sprintf("Error copying git diff output: %s", err.Error()))
		return "", err
	}
	err = diffCmd.Wait()
	if err != nil {
		logger.Error(fmt.Sprintf("Error waiting for git diff cmd to exit: %s", err.Error()))
		return "", err
	}
	return buf.String(), nil
}
