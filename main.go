package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

const timeFormats = `
Date formats:
(assumes local timezone if -0500 omitted)
default:        Mon Jul 3 17:18:43 2006 -0500
rfc2822:        Mon, 3 Jul 2006 17:18:43 -0500
iso8601:        2006-07-03 17:18:43 -0500
                2006-07-03 17:18:43 (local)
relative:       5.seconds.ago,
                2.years.3.months.ago,
                '6am yesterday'
`

type flagConfigInt struct {
	value int
	usage string
}

var flagsConfig = map[string]flagConfigInt{
	"commits":     {value: 10, usage: "Limit the number of commits to output."},
	"skipCommits": {value: 0, usage: "Skip number commits before starting to show the commit output."},
}

var (
	commits     = flag.Int("number", flagsConfig["commits"].value, flagsConfig["commits"].usage)
	skipCommits = flag.Int("skip", flagsConfig["skipCommits"].value, flagsConfig["skipCommits"].usage)
)

func init() {
	// example with short version for long flag
	flag.IntVar(commits, "n", flagsConfig["commits"].value, flagsConfig["commits"].usage)
}

func main() {
	flag.Parse()
	// fmt.Println("commits", *commits, "skipCommits", *skipCommits)
	cmd := exec.Command("git", "log", fmt.Sprintf("-n%d", *commits), "--format=format:%h@#@%ar@#@%s@#@%an@#@%d")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed: %s\n", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for i := 0; ; i++ {
		if scanner.Scan() == false {
			break
		}
		commit := parseGitLogLine(scanner.Text(), "@#@")
		fmt.Printf("[%d] %s\n", i, commit.color())
	}

	scanner = bufio.NewScanner(os.Stdin)
	fmt.Printf("Which previous commit? (q to quit) ")
	scanner.Scan()
	input := scanner.Text()
	rev, err := strconv.Atoi(input)
	if err != nil {
		if input == "q" {
			fmt.Println("Quitting")
			return
		}
		log.Fatalf("strconv.Atoi(input) failed: %s\n", err)
	}

	revSha := getLastNRevSha(rev)
	authorDate := getAuthorDate(rev)
	committerDate := getCommitterDate(rev)

	commitColor := color.New(color.FgYellow).PrintfFunc()
	commitColor("\ncommit %s\n", revSha)
	commitColor("GIT_AUTHOR_DATE=\"%s\"\n", authorDate)
	commitColor("GIT_COMMITTER_DATE=\"%s\"\n", committerDate)

	fmt.Println(timeFormats)
	// git filter-branch -f --env-filter \
	// 'if [ $GIT_COMMIT = f008f37282ce26ccf1e15a5fd3bf957371e71c77 ]
	//  then
	//      export GIT_AUTHOR_DATE="Thu Jul 26 9:43:44 2019 -0500"
	//      export GIT_COMMITTER_DATE="Thu Jul 26 9:43:44 2019 -0500"
	//  fi'

	var sameDate bool
	fmt.Printf("Are GIT_AUTHOR_DATE and GIT_COMMITTER_DATE the same? (y/n/q) ")
	scanner.Scan()
	input = strings.TrimSpace(scanner.Text())
	switch input {
	case "q":
		fmt.Println("Quitting")
		return
	case "n", "N":
		sameDate = false
	default:
		sameDate = true
	}

	fmt.Printf("new GIT_AUTHOR_DATE=")
	scanner.Scan()
	input = scanner.Text()
	if input == "q" {
		fmt.Println("Quitting")
		return
	}
	newAuthorDate := strings.TrimSpace(input)

	var newCommitterDate string
	if sameDate {
		newCommitterDate = newAuthorDate
	} else {
		fmt.Printf("new GIT_COMMITTER_DATE=")
		scanner.Scan()
		input = scanner.Text()
		if input == "q" {
			fmt.Println("Quitting")
			return
		}
		newCommitterDate = strings.TrimSpace(input)
	}

	envFilter := fmt.Sprintf("if [ $GIT_COMMIT = %s ]\nthen\nexport GIT_AUTHOR_DATE=\"%s\"\nexport GIT_COMMITTER_DATE=\"%s\"\nfi", revSha, newAuthorDate, newCommitterDate)
	cmd = exec.Command("git", "filter-branch", "-f", "--env-filter", envFilter)
	buf, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(buf))
		log.Fatalf("cmd.Run() failed: %v\n", err)
	}
	fmt.Println(string(bytes.TrimRight(buf, "\n")))

	// git update-ref -d refs/original/refs/heads/master
	fmt.Print("\nRemove old refs using\n\n" +
		"\tgit update-ref -d refs/original/refs/heads/master\n\n" +
		"See https://stackoverflow.com/a/7654880 for details\n")
}

func getAuthorDate(rev int) string {
	cmd := exec.Command("git", "show", "-s", "--format=%ai", fmt.Sprintf("HEAD~%d", rev))
	buf, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed: %s\n", err)
	}
	return string(bytes.TrimRight(buf, "\n"))
}

func getCommitterDate(rev int) string {
	cmd := exec.Command("git", "show", "-s", "--format=%ci", fmt.Sprintf("HEAD~%d", rev))
	buf, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed: %s\n", err)
	}
	return string(bytes.TrimRight(buf, "\n"))
}

func getLastNRevSha(rev int) string {
	cmd := exec.Command("git", "rev-parse", fmt.Sprintf("HEAD~%d", rev))
	buf, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed: %s\n", err)
	}
	return string(bytes.TrimRight(buf, "\n"))
}

type gitLogLine struct {
	Sha     string
	Date    string
	Message string
	Author  string
	Tags    string
}

func parseGitLogLine(s, sep string) gitLogLine {
	x := strings.Split(s, sep)
	return gitLogLine{Sha: x[0], Date: x[1], Message: x[2], Author: x[3], Tags: x[4]}
}

func (line gitLogLine) color() string {
	shaColor := color.New(color.FgBlue, color.Bold).SprintFunc()
	timeColor := color.New(color.FgGreen, color.Bold).SprintFunc()
	authorColor := color.New(color.Faint).SprintFunc()
	tagColor := color.New(color.FgYellow, color.Bold).SprintFunc()

	return fmt.Sprintf("%s - (%s) %s - %s%s", shaColor(line.Sha), timeColor(line.Date), line.Message, authorColor(line.Author), tagColor(line.Tags))
}
