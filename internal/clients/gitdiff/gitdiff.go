package gitdiff

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

func printFiles(dir string) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("Error in accessing directory:%s", err.Error())
	}

	for _, file := range fileInfos {
		log.Printf(file.Name())
	}
}

/*
func (gd *GitDiff) GetDiff() (string, error) {
	fakeDiff :=
		`diff --git a/.github/workflows/nightfalldlp.yml b/.github/workflows/nightfalldlp.yml
new file mode 100644
index 0000000..9b8b60b
--- /dev/null
+++ b/.github/workflows/nightfalldlp.yml
@@ -0,0 +1,21 @@
+name: nightfalldlp
+on:
+  push:
+    branches:
+      - master
+  pull_request:
+jobs:
+  nightfalldlp:
+    name: nightfalldlp
+    runs-on: self-hosted
+    steps:
+      - name: Checkout Repo Action
+        uses: actions/checkout@v2
+
+      - name: nightfallDLP action step
+        uses: nightfallai/nightfall_dlp_action@TestGithubClient
+        env:
+          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
+          NIGHTFALL_API_KEY: ${{ secrets.NIGHTFALL_API_KEY }}
+          EVENT_BEFORE: ${{ github.event.before }}
+          BASE_URL: https://ec2-54-184-124-231.us-west-2.compute.amazonaws.com
diff --git a/.gitignore b/.gitignore
new file mode 100644
index 0000000..9f11b75
--- /dev/null
+++ b/.gitignore
@@ -0,0 +1 @@
+.idea/
diff --git a/.nightfalldlp/config.json b/.nightfalldlp/config.json
new file mode 100644
index 0000000..35b011d
--- /dev/null
+++ b/.nightfalldlp/config.json
@@ -0,0 +1,43 @@
+{
+  "conditions": [
+    {
+      "detector": {
+        "detectorType": "NIGHTFALL_DETECTOR",
+        "nightfallDetector": "CREDIT_CARD_NUMBER",
+        "displayName": "cc num"
+      },
+      "minNumFindings": 1,
+      "minConfidence": "LIKELY"
+    },
+    {
+      "detector": {
+        "detectorType": "NIGHTFALL_DETECTOR",
+        "nightfallDetector": "PHONE_NUMBER",
+        "displayName": "phone num"
+      },
+      "minNumFindings": 1,
+      "minConfidence": "LIKELY"
+    },
+    {
+      "detector": {
+        "detectorType": "NIGHTFALL_DETECTOR",
+        "nightfallDetector": "API_KEY",
+        "displayName": "api key"
+      },
+      "minNumFindings": 1,
+      "minConfidence": "LIKELY"
+    },
+    {
+      "detector": {
+        "detectorType": "NIGHTFALL_DETECTOR",
+        "nightfallDetector": "CRYPTOGRAPHIC_KEY",
+        "displayName": "crypto key"
+      },
+      "minNumFindings": 1,
+      "minConfidence": "LIKELY"
+    }
+  ],
+  "maxNumberConcurrentRoutines": 5,
+  "tokenExclusionList": ["4916-6734-7572-5015"],
+  "fileInclusionList": ["*"]
+}
diff --git a/test/findings.txt b/test/findings.txt
new file mode 100644
index 0000000..86dd0a4
--- /dev/null
+++ b/test/findings.txt
@@ -0,0 +1,2 @@
+my phone number is 301-231-5329
+my cc is 4242-4242-4242-4242
`

	return fakeDiff, nil
}*/

func printFile(path string) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("err reading file: %s", err.Error())
	}
	log.Printf("git .config contents: %s", string(fileBytes))
}

// GetDiff uses the command line to compute the diff
func (gd *GitDiff) GetDiff() (string, error) {
	origWd, err := os.Getwd()
	printFiles(origWd)
	printFiles(fmt.Sprintf("%s%s", origWd, "/.git"))
	printFile(fmt.Sprintf("%s%s", origWd, "/.git/config"))
	log.Printf("orig work dir: %s", origWd)
	if err != nil {
		return "", err
	}
	err = os.Chdir(gd.WorkDir)
	//err = os.Chdir("~/actions-runner/_work/TestRepo2")
	if err != nil {
		return "", err
	}
	log.Printf("changed work dir: %s", gd.WorkDir)
	logger := log.New(os.Stdout, "", 0)
	logCmd := exec.Command("git", "log")
	logBytes, err := logCmd.CombinedOutput()
	if err != nil {
		log.Printf("git log err:%s", err.Error())
	}
	log.Printf("git log bytes: %s\n", string(logBytes))
	var diffCmd *exec.Cmd
	/*pwdCmd := exec.Command("pwd")
	err = pwdCmd.Run()
	if err != nil {
		logger.Printf("err running pwd command: %s", pwdCmd.String())
	}*/
	switch {
	case gd.BaseBranch != "":
		gitCmd := exec.Command("git", "-c", "http.sslVerify=false", "fetch", "origin", gd.BaseBranch, "--depth=1")
		logger.Printf("running git command PR: %s", gitCmd.String())
		// PR event so get diff between base branch and current commit SHA
		// err = gitCmd.Run()
		gitFetchBytes, err := gitCmd.CombinedOutput()
		if err != nil {
			logger.Printf("running git fetch origin: %s", err.Error())
			logger.Printf("git fetch output: %s", string(gitFetchBytes))
			return "", err
		}
		log.Printf("git fetch output: %s\n", string(gitFetchBytes))
		diffCmd = exec.Command("git", "diff", fmt.Sprintf("origin/%s", gd.BaseBranch))
	case gd.BaseSHA == "" || gd.BaseSHA == unknownCommitHash:
		// PUSH event for new branch so use git show to get the diff of the most recent commit
		err = exec.Command("git", "fetch", "origin", gd.Head, "--depth=2").Run()
		if err != nil {
			return "", err
		}
		diffCmd = exec.Command("git", "show", gd.Head, "--format=")
	default:
		// PUSH event where last commit action ran on exists
		// use current commit SHA and previous action run commit SHA for diff
		err = exec.Command("git", "fetch", "origin", gd.BaseSHA, "--depth=1").Run()
		if err != nil {
			return "", err
		}
		diffCmd = exec.Command("git", "diff", gd.BaseSHA, gd.Head)
	}

	reader, err := diffCmd.StdoutPipe()
	if err != nil {
		logger.Printf("running diff stdoutpipe: %s", err.Error())
		return "", err
	}
	defer reader.Close()

	err = diffCmd.Start()
	if err != nil {
		logger.Printf("running diff start: %s", err.Error())
		return "", nil
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		logger.Printf("running diff copy buf: %s", err.Error())
		return "", err
	}
	err = diffCmd.Wait()
	if err != nil {
		logger.Printf("running diff cmd wait: %s", err.Error())
		return "", err
	}
	return buf.String(), nil
}
