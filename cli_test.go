package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHelpOutputAndExit(t *testing.T) {
	io := NewMemIo(map[string]string{})
	status := Run([]string{"--help"}, io)
	if status != 0 {
		t.Errorf("Run(--help) status = %d, want 0", status)
	}
	if got, want := io.Stdout, "usage: skillplumb check DIR [--max-desc N] [--max-lines N] [--ignore CODE]... [--strict]\n       skillplumb index DIR\n       skillplumb --help\n"; got != want {
		t.Errorf("Run(--help) stdout = %q, want %q", got, want)
	}
	if io.Stderr != "" {
		t.Errorf("Run(--help) stderr = %q, want empty", io.Stderr)
	}
}

func workedExampleIo() *MemIo {
	return NewMemIo(map[string]string{
		"skills/README.md":       "project notes",
		"skills/.cache/SKILL.md": "x",
		"skills/alpha/rules.md":  "r",
		"skills/alpha/SKILL.md":  "---\nname: alpha\ndescription: Reviews a pull request.\nlicense: MIT\n---\n# Alpha\nSee [the rules](rules.md) and [the site](https://example.org).\n",
		"skills/beta/notes/a.md": "r",
		"skills/beta/SKILL.md":   "---\nname: Beta\ndescription: reviews  a PULL request.\n---\nSee [notes](notes/a.md), [todo](todo.md) and [up](../alpha/rules.md).\n",
		"skills/delta/SKILL.md":  "---\n---\n",
		"skills/gamma/notes.txt": "notes",
		"skills/zeta/SKILL.md":   "# No front matter\n",
	})
}

func TestWorkedCheckOutputAndExit(t *testing.T) {
	io := workedExampleIo()
	status := Run([]string{"check", "skills"}, io)
	if status != 1 {
		t.Errorf("Run(check skills) status = %d, want 1", status)
	}
	want := "beta/SKILL.md:2: E06 name Beta is not valid\n" +
		"beta/SKILL.md:3: W04 description is the same as in skill alpha\n" +
		"beta/SKILL.md:5: E09 link todo.md does not exist\n" +
		"beta/SKILL.md:5: E10 link ../alpha/rules.md leaves the skill folder\n" +
		"delta/SKILL.md:1: E06 name is missing\n" +
		"delta/SKILL.md:1: E07 description is missing\n" +
		"delta/SKILL.md:2: W03 body is empty\n" +
		"gamma/SKILL.md:1: E01 SKILL.md is missing\n" +
		"zeta/SKILL.md:1: E03 front matter is missing or not closed\n" +
		"skillplumb: 5 skills, 7 errors, 2 warnings\n"
	if io.Stdout != want {
		t.Errorf("Run(check skills) stdout = %q, want %q", io.Stdout, want)
	}
	if io.Stderr != "" {
		t.Errorf("Run(check skills) stderr = %q, want empty", io.Stderr)
	}
}

func TestWorkedIndexOutputAndExit(t *testing.T) {
	io := workedExampleIo()
	status := Run([]string{"index", "skills"}, io)
	if status != 0 {
		t.Errorf("Run(index skills) status = %d, want 0", status)
	}
	want := "skillplumb: 5 skills\n" +
		"SKILL  LINES  LINKS  DESC\n" +
		"alpha  2      2      23\n" +
		"beta   1      3      24\n" +
		"delta  0      0      -\n" +
		"gamma  -      -      -\n" +
		"zeta   -      -      -\n"
	if io.Stdout != want {
		t.Errorf("Run(index skills) stdout = %q, want %q", io.Stdout, want)
	}
	if io.Stderr != "" {
		t.Errorf("Run(index skills) stderr = %q, want empty", io.Stderr)
	}
}

func TestCheckExitStatuses(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		file   string
		status int
		stdout string
	}{
		{
			name:   "valid skill",
			args:   []string{"check", "d"},
			file:   "---\nname: a\ndescription: valid\n---\ncontent\n",
			status: 0,
			stdout: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
		},
		{
			name:   "warning without strict",
			args:   []string{"check", "d"},
			file:   "---\nname: a\ndescription: valid\n---\n",
			status: 0,
			stdout: "a/SKILL.md:4: W03 body is empty\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
		},
		{
			name:   "warning with strict",
			args:   []string{"check", "d", "--strict"},
			file:   "---\nname: a\ndescription: valid\n---\n",
			status: 1,
			stdout: "a/SKILL.md:4: W03 body is empty\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
		},
		{
			name:   "error",
			args:   []string{"check", "d"},
			file:   "---\nname: A\ndescription: valid\n---\nx\n",
			status: 1,
			stdout: "a/SKILL.md:2: E06 name A is not valid\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
		},
	}
	for _, test := range cases {
		io := NewMemIo(map[string]string{
			"d/a/SKILL.md": test.file,
		})
		status := Run(test.args, io)
		if status != test.status {
			t.Errorf("%s: Run(%q) status = %d, want %d", test.name, test.args, status, test.status)
		}
		if io.Stdout != test.stdout {
			t.Errorf("%s: Run(%q) stdout = %q, want %q", test.name, test.args, io.Stdout, test.stdout)
		}
		if io.Stderr != "" {
			t.Errorf("%s: Run(%q) stderr = %q, want empty", test.name, test.args, io.Stderr)
		}
	}
}

func TestProcessCRLFBoundary(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temporaryDirectory := t.TempDir()
	binary := filepath.Join(temporaryDirectory, "skillplumb")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	skillDir := filepath.Join(temporaryDirectory, "alpha")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	content := []byte("---\r\nname: alpha\r\ndescription: Does a thing.\r\n---\r\nSee [x](missing.md)\r\n")
	if err := os.WriteFile(skillFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"check", temporaryDirectory}
	run := exec.Command(binary, args...)
	run.Stdout = &stdout
	run.Stderr = &stderr
	err = run.Run()
	status := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("program failed to run: %v", err)
		}
		status = exitErr.ExitCode()
	}
	if status != 1 {
		t.Errorf("program exit status = %d, want 1", status)
	}
	if got, want := stderr.String(), ""; got != want {
		t.Errorf("program stderr = %q, want %q", got, want)
	}
	want := "alpha/SKILL.md:5: E09 link missing.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n"
	if got := stdout.String(); got != want {
		t.Errorf("program stdout = %q, want %q", got, want)
	}
}
