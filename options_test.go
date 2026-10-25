package main

import "testing"

func TestAC012CommandErrors(t *testing.T) {
	cases := []struct {
		args   []string
		stderr string
	}{
		{[]string{}, "skillplumb: missing command (check or index)\n"},
		{[]string{"--help", "x"}, "skillplumb: unknown command --help\n"},
		{[]string{"check"}, "skillplumb: missing DIR\n"},
		{[]string{"check", "d", "--max-desc"}, "skillplumb: --max-desc needs a value\n"},
		{[]string{"check", "d", "--max-desc", "0"}, "skillplumb: --max-desc needs a whole number of 1 or more, got 0\n"},
		{[]string{"check", "d", "--max-desc", "12x"}, "skillplumb: --max-desc needs a whole number of 1 or more, got 12x\n"},
		{[]string{"check", "d", "--max-lines", "\t"}, "skillplumb: --max-lines needs a whole number of 1 or more, got ?\n"},
		{[]string{"check", "d", "--max-desc", "5", "--max-desc", "6"}, "skillplumb: --max-desc given twice\n"},
		{[]string{"check", "d", "--ignore", "E00", "--ignore", "e01"}, "skillplumb: unknown code E00\n"},
		{[]string{"check", "d", "--max-desc=5"}, "skillplumb: unknown option --max-desc=5\n"},
		{[]string{"index", "d", "--strict"}, "skillplumb: unknown option --strict\n"},
		{[]string{"check", "d", "e"}, "skillplumb: unexpected argument e\n"},
		{[]string{"check", "nope"}, "skillplumb: nope: not a directory\n"},
		{[]string{"check", "a\tb"}, "skillplumb: a?b: not a directory\n"},
		{[]string{"check", "empty"}, "skillplumb: empty: no skills found\n"},
		{[]string{"check", "d", "--max-desc", "x", "--ignore", "zz"}, "skillplumb: --max-desc needs a whole number of 1 or more, got x\n"},
		{[]string{"check", ""}, "skillplumb: (empty): not a directory\n"},
		{[]string{"check", "d", "--ignore", ""}, "skillplumb: unknown code (empty)\n"},
		{[]string{"", "d"}, "skillplumb: unknown command (empty)\n"},
	}

	for _, test := range cases {
		io := NewMemIo(map[string]string{
			"d/a/SKILL.md": "---\nname: a\ndescription: valid\n---\nx\n",
			"d/f.txt":      "file",
			"empty/":       "",
		})
		status := Run(test.args, io)
		if status != 2 {
			t.Errorf("Run(%q) status = %d, want 2", test.args, status)
		}
		if io.Stdout != "" {
			t.Errorf("Run(%q) stdout = %q, want empty", test.args, io.Stdout)
		}
		if io.Stderr != test.stderr {
			t.Errorf("Run(%q) stderr = %q, want %q", test.args, io.Stderr, test.stderr)
		}
	}
}

func optionsTestIo() *MemIo {
	return NewMemIo(map[string]string{
		"d/a/SKILL.md":     "---\nname: a\ndescription: valid\n---\nx\n",
		"-/a/SKILL.md":     "---\nname: a\ndescription: valid\n---\nx\n",
		"-x/a/SKILL.md":    "---\nname: a\ndescription: valid\n---\nx\n",
		"warn/a/SKILL.md":  "---\nname: a\ndescription: valid\n---\n",
		"mixed/a/":         "",
		"mixed/b/SKILL.md": "---\nname: b\ndescription: valid\n---\n",
	})
}

func TestAC013OptionPrecedenceAndDashOperands(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		status int
		stdout string
		stderr string
	}{
		{"option before DIR and leading zeroes", []string{"check", "--max-desc", "000005", "d"}, 0, "skillplumb: 1 skills, 0 errors, 0 warnings\n", ""},
		{"six digit max-lines", []string{"check", "d", "--max-lines", "999999"}, 0, "skillplumb: 1 skills, 0 errors, 0 warnings\n", ""},
		{"seven digit max-lines", []string{"check", "d", "--max-lines", "1000000"}, 2, "", "skillplumb: --max-lines needs a whole number of 1 or more, got 1000000\n"},
		{"strict given twice", []string{"check", "d", "--strict", "--strict"}, 2, "", "skillplumb: --strict given twice\n"},
		{"option-looking value is consumed", []string{"check", "d", "--max-desc", "--strict"}, 2, "", "skillplumb: --max-desc needs a whole number of 1 or more, got --strict\n"},
		{"max-desc semantic error precedes max-lines", []string{"check", "d", "--max-lines", "x", "--max-desc", "y"}, 2, "", "skillplumb: --max-desc needs a whole number of 1 or more, got y\n"},
		{"structural error precedes semantic error", []string{"check", "d", "--max-desc", "x", "--wat"}, 2, "", "skillplumb: unknown option --wat\n"},
		{"missing DIR precedes semantic error", []string{"check", "--max-desc", "x"}, 2, "", "skillplumb: missing DIR\n"},
		{"single dash is operand", []string{"check", "-"}, 0, "skillplumb: 1 skills, 0 errors, 0 warnings\n", ""},
		{"dash letter is operand", []string{"check", "-x"}, 0, "skillplumb: 1 skills, 0 errors, 0 warnings\n", ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			io := optionsTestIo()
			status := Run(test.args, io)
			if status != test.status {
				t.Errorf("Run(%q) status = %d, want %d", test.args, status, test.status)
			}
			if io.Stdout != test.stdout {
				t.Errorf("Run(%q) stdout = %q, want %q", test.args, io.Stdout, test.stdout)
			}
			if io.Stderr != test.stderr {
				t.Errorf("Run(%q) stderr = %q, want %q", test.args, io.Stderr, test.stderr)
			}
		})
	}
}

func TestAC013EveryIgnoreCodeIsAccepted(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"E01", "E01"},
		{"E02", "E02"},
		{"E03", "E03"},
		{"E04", "E04"},
		{"E05", "E05"},
		{"E06", "E06"},
		{"E07", "E07"},
		{"E08", "E08"},
		{"E09", "E09"},
		{"E10", "E10"},
		{"W01", "W01"},
		{"W02", "W02"},
		{"W03", "W03"},
		{"W04", "W04"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			io := optionsTestIo()
			status := Run([]string{"check", "d", "--ignore", test.code}, io)
			if status != 0 {
				t.Errorf("Run with --ignore %s status = %d, want 0", test.code, status)
			}
			if io.Stdout != "skillplumb: 1 skills, 0 errors, 0 warnings\n" {
				t.Errorf("Run with --ignore %s stdout = %q", test.code, io.Stdout)
			}
			if io.Stderr != "" {
				t.Errorf("Run with --ignore %s stderr = %q, want empty", test.code, io.Stderr)
			}
		})
	}
}

func TestAC013RejectedIgnoreSpellings(t *testing.T) {
	cases := []struct {
		code   string
		stderr string
	}{
		{"e01", "skillplumb: unknown code e01\n"},
		{"E11", "skillplumb: unknown code E11\n"},
		{"W05", "skillplumb: unknown code W05\n"},
		{"E001", "skillplumb: unknown code E001\n"},
	}
	for _, test := range cases {
		io := optionsTestIo()
		status := Run([]string{"check", "d", "--ignore", test.code}, io)
		if status != 2 {
			t.Errorf("Run with --ignore %s status = %d, want 2", test.code, status)
		}
		if io.Stdout != "" {
			t.Errorf("Run with --ignore %s stdout = %q, want empty", test.code, io.Stdout)
		}
		if io.Stderr != test.stderr {
			t.Errorf("Run with --ignore %s stderr = %q, want %q", test.code, io.Stderr, test.stderr)
		}
	}
}

func TestAC013FilteringAndExitStatuses(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		status int
		stdout string
	}{
		{"clean check exits zero", []string{"check", "d"}, 0, "skillplumb: 1 skills, 0 errors, 0 warnings\n"},
		{"warning-only check exits zero", []string{"check", "warn"}, 0, "a/SKILL.md:4: W03 body is empty\nskillplumb: 1 skills, 0 errors, 1 warnings\n"},
		{"strict warning-only check exits one", []string{"check", "warn", "--strict"}, 1, "a/SKILL.md:4: W03 body is empty\nskillplumb: 1 skills, 0 errors, 1 warnings\n"},
		{"ignored strict warning exits zero", []string{"check", "warn", "--strict", "--ignore", "W03"}, 0, "skillplumb: 1 skills, 0 errors, 0 warnings\n"},
		{"ignore error leaves warning", []string{"check", "mixed", "--ignore", "E01"}, 0, "b/SKILL.md:4: W03 body is empty\nskillplumb: 2 skills, 0 errors, 1 warnings\n"},
		{"ignore warning leaves error", []string{"check", "mixed", "--ignore", "W03"}, 1, "a/SKILL.md:1: E01 SKILL.md is missing\nskillplumb: 2 skills, 1 errors, 0 warnings\n"},
		{"repeated ignores filter findings", []string{"check", "mixed", "--ignore", "E01", "--ignore", "W03", "--ignore", "E01"}, 0, "skillplumb: 2 skills, 0 errors, 0 warnings\n"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			io := optionsTestIo()
			status := Run(test.args, io)
			if status != test.status {
				t.Errorf("Run(%q) status = %d, want %d", test.args, status, test.status)
			}
			if io.Stdout != test.stdout {
				t.Errorf("Run(%q) stdout = %q, want %q", test.args, io.Stdout, test.stdout)
			}
			if io.Stderr != "" {
				t.Errorf("Run(%q) stderr = %q, want empty", test.args, io.Stderr)
			}
		})
	}
}
