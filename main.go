package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
)

const (
	NOTHING = 0
	PATCH   = 1
	MINOR   = 2
	MAJOR   = 3
)

var (
	PATCH_KEYWORDS = []string{"fix"}
	MINOR_KEYWORDS = []string{"feat"}
	MAJOR_KEYWORDS = []string{}
	NOTE_KEYWORDS  = []string{"BREAKING CHANGE", "BREAKING CHANGES"}
)

func main() {
	toPtr := flag.String("to", "HEAD", "commitish to diff against")
	releaseBranchPtr := flag.String("release-branch", "master", "the release branch")
	flag.Parse()

	lastVersion, err := getLastVersion(*releaseBranchPtr)
	if err != nil {
		fmt.Printf("Could not determine last tag, %v\n", err)
		os.Exit(1)
	}

	commits, err := getCommits(*toPtr, fmt.Sprintf("v%s", lastVersion))
	if err != nil {
		fmt.Printf("Could not get commits, %v\n", err)
		os.Exit(1)
	}

	messages, err := getMessages(commits)
	if err != nil {
		fmt.Printf("Error parsing commit messages, %v\n", err)
		os.Exit(1)
	}

	bump, err := getBump(messages)
	if err != nil {
		fmt.Printf("Error determining bump, %v\n", err)
		os.Exit(1)
	}

	version := lastVersion

	switch bump {
	case PATCH:
		version.BumpPatch()
	case MINOR:
		version.BumpMinor()
	case MAJOR:
		version.BumpMajor()
	default:
		os.Exit(0)
	}

	fmt.Printf("v%s\n", version)
}

func getBump(messages []string) (int, error) {
	r, err := regexp.Compile("[a-zA-Z]+")

	if err != nil {
		return 0, err
	}

	action := NOTHING

	for _, message := range messages {
		lines := strings.Split(message, "\n")
		subject := lines[0]
		note := strings.Join(lines[1:], " ")

		for _, keyword := range MAJOR_KEYWORDS {
			if r.FindString(subject) == keyword {
				action = MAJOR
				break
			}
		}

		for _, keyword := range MINOR_KEYWORDS {
			if r.FindString(subject) == keyword {
				action = MINOR
				break
			}
		}

		for _, keyword := range PATCH_KEYWORDS {
			if r.FindString(subject) == keyword {
				action = PATCH
				break
			}
		}

		for _, noteKeyword := range NOTE_KEYWORDS {
			if strings.Contains(note, noteKeyword) == true {
				action = MAJOR
				break
			}
		}
	}

	return action, nil
}

func getMessages(commits []Hash) ([]string, error) {
	var messages []string

	for _, c := range commits {
		output, err := exec.Command("git", "show", "--pretty=format:%B", string(c)).CombinedOutput()

		if err == nil {
			messages = append(messages, string(output))
		} else {
			return nil, err
		}
	}

	return messages, nil
}

type Hash string

func getCommits(head string, last string) ([]Hash, error) {
	output, err := exec.Command("git", "cherry", last, head).CombinedOutput()

	if err != nil {
		return nil, err
	}

	var commits []Hash

	lines := strings.Split(string(output), "\n")

	for _, h := range lines {
		if h != "" {
			hash := Hash(strings.TrimLeft(h, "+ "))
			commits = append(commits, hash)
		}
	}

	return commits, nil
}

func getLastVersion(releaseBranch string) (*semver.Version, error) {
	output, err := exec.Command("git", "tag", "--merged", releaseBranch).CombinedOutput()
	lines := strings.Split(string(output), "\n")

	if err != nil {
		return nil, err
	}

	var versions []*semver.Version

	lineR, err := regexp.Compile("\\d+\\.\\d+\\.\\d+")

	for _, line := range lines {
		tag := lineR.FindString(line)

		if tag != "" {
			v, err := semver.NewVersion(tag)

			if err == nil {
				versions = append(versions, v)
			}
		}
	}

	semver.Sort(versions)

	return versions[len(versions)-1], nil
}
