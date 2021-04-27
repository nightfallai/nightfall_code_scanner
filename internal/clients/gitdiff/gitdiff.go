package gitdiff

import (
	"bufio"
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
		logger.Info("Getting Diff between Base Branch and Current Commit SHA")
		// PR event so get diff between base branch and current commit SHA
		fetchCmd := exec.Command("git", "fetch", "origin", gd.BaseBranch, "--depth=1")
		fetchCmd.Env = os.Environ()
		fetchCmd.Env = append(fetchCmd.Env,
			"GIT_DISCOVERY_ACROSS_FILESYSTEM=1",
			"GIT_TRACE=true",
			"GIT_TRACE=true",
			"GIT_CURL_VERBOSE=true",
			"GIT_SSH_COMMAND=\"ssh -vvv\"\t",
			"GIT_TRACE_SHALLOW=true")
		reader, err := fetchCmd.StdoutPipe()
		errReader, err := fetchCmd.StderrPipe()
		if err != nil {
			logger.Error(fmt.Sprintf("Error piping git fetch cmd: %s", err.Error()))
			return "", err
		}
		defer reader.Close()
		defer errReader.Close()
		multi := io.MultiReader(reader, errReader)
		in := bufio.NewScanner(multi)
		err = fetchCmd.Start()
		if err != nil {
			logger.Error(fmt.Sprintf("Error starting git fetch cmd: %s", err.Error()))
			return "", nil
		}
		for in.Scan() {
			logger.Info(in.Text()) // write each line to your log, or anything you need
		}
		if err := in.Err(); err != nil {
			log.Printf("error: %s", err)
		}
		buf := new(strings.Builder)
		_, err = io.Copy(buf, reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Error copying git fetch output: %s", err.Error()))
			return "", err
		}
		err = fetchCmd.Wait()
		if err != nil {
			logger.Error(fmt.Sprintf("Error waiting for git fetch cmd to exit: %s", err.Error()))
			return "", err
		}
		logger.Info(fmt.Sprintf("Fetch Origin Logs: %s", buf.String()))
		/*err = exec.Command("git", "-c", "http.sslVerify=false", "fetch", "origin", gd.BaseBranch, "--depth=1").Run()
		if err != nil {
			logger.Error(fmt.Sprintf("Error getting diff between Base Branch and Current Commit SHA: %s", err.Error()))
			return "", err
		}*/
		/*err = fetchCmd.Run()
		if err != nil {
			logger.Error(fmt.Sprintf("Error getting diff between Base Branch and Current Commit SHA: %s", err.Error()))
			return "", err
		}*/
		diffCmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s", gd.BaseBranch))
	case gd.BaseSHA == "" || gd.BaseSHA == unknownCommitHash:
		logger.Info("Getting Diff for new branch push event")
		// PUSH event for new branch so use git show to get the diff of the most recent commit
		err = exec.Command("git", "fetch", "origin", gd.Head, "--depth=2").Run()
		if err != nil {
			logger.Error(fmt.Sprintf("Error getting diff in new branch push event: %s", err.Error()))
			return "", err
		}
		diffCmd = exec.Command("git", "show", gd.Head, "--format=")
	default:
		// PUSH event where last commit action ran on exists
		// use current commit SHA and previous action run commit SHA for diff
		logger.Info("Getting Diff for new push event")
		fetchCmd := exec.Command("git", "fetch", "origin", gd.BaseSHA, "--depth=1")
		fetchCmd.Env = os.Environ()
		fetchCmd.Env = append(fetchCmd.Env,
			"GIT_DISCOVERY_ACROSS_FILESYSTEM=1",
			"GIT_TRACE=true",
			"GIT_CURL_VERBOSE=true",
			"GIT_SSH_COMMAND=\"ssh -vvv\"\t",
			"GIT_TRACE_SHALLOW=true")
		reader, err := fetchCmd.StdoutPipe()
		errReader, err := fetchCmd.StderrPipe()
		if err != nil {
			logger.Error(fmt.Sprintf("Error piping git fetch cmd: %s", err.Error()))
			return "", err
		}
		defer reader.Close()
		defer errReader.Close()
		multi := io.MultiReader(reader, errReader)
		in := bufio.NewScanner(multi)
		err = fetchCmd.Start()
		if err != nil {
			logger.Error(fmt.Sprintf("Error starting git fetch cmd: %s", err.Error()))
			return "", nil
		}
		for in.Scan() {
			logger.Info(in.Text()) // write each line to your log, or anything you need
		}
		if err := in.Err(); err != nil {
			log.Printf("error: %s", err)
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
