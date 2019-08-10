package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	switch os.Getenv("GO_TEST_MODE") {
	case "":
		// Normal test mode
		os.Exit(m.Run())

	case "echo":
		iargs := []interface{}{}
		for _, s := range os.Args[1:] {
			iargs = append(iargs, s)
		}
		fmt.Println(iargs...)

	case "getLastNRevHash":
		fmt.Printf(testHash)
	}
}

// example test from os.exec test suite: https://golang.org/src/os/exec/exec_test.go
func TestEcho(t *testing.T) {
	cmd := exec.Command(os.Args[0], "hello", "world")
	cmd.Env = []string{"GO_TEST_MODE=echo"}
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("echo: %v", err)
	}
	if g, e := string(output), "hello world\n"; g != e {
		t.Errorf("echo: want %q, got %q", e, g)
	}
}

const testHash = "0cfd8742049972a90b68021353add7a3b5134316"

func TestGetLastNRevHash(t *testing.T) {
	cmd := exec.Command(os.Args[0], "git", "rev-parse", fmt.Sprintf("HEAD~%d", 2))
	cmd.Env = []string{"GO_TEST_MODE=getLastNRevHash"}
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("getLastNRevHash: %v", err)
	}
	if g, e := string(output), testHash; g != e {
		t.Errorf("getLastNRevHash: want %q, got %q", e, g)
	}
}

func TestParseGitLogLine(t *testing.T) {
	testLine, testSep := "1234567#(date)#Description#Name#(HEAD -> master, origin/master)", "#"
	got := parseGitLogLine(testLine, testSep)
	want := &gitLogLine{"1234567", "(date)", "Description", "Name", "(HEAD -> master, origin/master)"}
	if *got != *want {
		t.Errorf("splitGitLogLine(%v, %v) => %q, want %q",
			testLine, testSep, got, want)
	}
}
