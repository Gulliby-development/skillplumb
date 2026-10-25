package main

import "testing"

func TestLiteralLinkScanningAndResolution(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		body      string
		checkWant string
		indexWant string
	}{
		{
			name:      "existing file",
			status:    0,
			body:      "[a](ref.md)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "dot and slash normalization",
			status:    0,
			body:      "[a](./ref.md) [b](sub/f.md) [c](sub/) [d](sub) [e](.)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      5      1\n",
		},
		{
			name:      "fragment is removed for lookup",
			status:    0,
			body:      "[a](ref.md#top)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "external and ignored targets are counted",
			status:    0,
			body:      "[a](#top) [b](https://x.org/y) [c](/abs) [d](c:d) [e]() [f](g h)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      4      1\n",
		},
		{
			name:      "missing target",
			status:    1,
			body:      "[a](nope.md)",
			checkWant: "s/SKILL.md:5: E09 link nope.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "case-sensitive target",
			status:    1,
			body:      "[a](Ref.md)",
			checkWant: "s/SKILL.md:5: E09 link Ref.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "file cannot be an intermediate directory",
			status:    1,
			body:      "[a](sub/f.md/g)",
			checkWant: "s/SKILL.md:5: E09 link sub/f.md/g does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "parent traversal leaves skill",
			status:    1,
			body:      "[a](../x)",
			checkWant: "s/SKILL.md:5: E10 link ../x leaves the skill folder\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "normalized traversal leaves skill",
			status:    1,
			body:      "[a](sub/../../x)",
			checkWant: "s/SKILL.md:5: E10 link sub/../../x leaves the skill folder\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "parent target leaves skill",
			status:    1,
			body:      "[a](..)",
			checkWant: "s/SKILL.md:5: E10 link .. leaves the skill folder\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "consecutive links continue after existing target",
			status:    1,
			body:      "[a](ref.md)[b](nope.md)",
			checkWant: "s/SKILL.md:5: E09 link nope.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      2      1\n",
		},
		{
			name:      "same-line findings retain target order",
			status:    1,
			body:      "[a](n1.md) and [b](n2.md)",
			checkWant: "s/SKILL.md:5: E09 link n1.md does not exist\ns/SKILL.md:5: E09 link n2.md does not exist\nskillplumb: 1 skills, 2 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      2      1\n",
		},
		{
			name:      "missing closing parenthesis stops the line",
			status:    0,
			body:      "[a](nope.md",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      1\n",
		},
		{
			name:      "inline code is scanned",
			status:    1,
			body:      "`[a](nope.md)`",
			checkWant: "s/SKILL.md:5: E09 link nope.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "first marker determines target",
			status:    1,
			body:      "[a](b](ref.md)",
			checkWant: "s/SKILL.md:5: E09 link b](ref.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "leading tab does not prevent scanning",
			status:    1,
			body:      "\t[a](nope.md)",
			checkWant: "s/SKILL.md:5: E09 link nope.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "fenced block is skipped but later line is scanned",
			status:    1,
			body:      "\u0060\u0060\u0060\n[a](nope.md)\n\u0060\u0060\u0060\n[b](nope.md)",
			checkWant: "s/SKILL.md:8: E09 link nope.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      4      1      1\n",
		},
		{
			name:      "fence lines with suffix toggle and repeated fences",
			status:    0,
			body:      "``` x\n[a](nope.md)\n```\n```\n[b](nope.md)\n```",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      6      0      1\n",
		},
		{
			name:      "unclosed fenced block suppresses the rest",
			status:    0,
			body:      "```\n[a](nope.md)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      2      0      1\n",
		},
		{
			name:      "tab inside target is ignored",
			status:    0,
			body:      "[a](nope\t.md)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      1\n",
		},
		{
			name:      "image syntax is scanned as a link",
			status:    1,
			body:      "![image](missing.png)",
			checkWant: "s/SKILL.md:5: E09 link missing.png does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "angle brackets do not change scanning",
			status:    1,
			body:      "<[a](missing.md)>",
			checkWant: "s/SKILL.md:5: E09 link missing.md does not exist\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "percent escape is looked up literally",
			status:    0,
			body:      "[a](percent%20name.md)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "percent token is a literal filename",
			status:    0,
			body:      "[a](%20)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
		{
			name:      "other entry can be a final target",
			status:    0,
			body:      "[a](special)",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      1      1\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			checkIO := newLinkTestMem(testCase.body)
			checkStatus := Run([]string{"check", "d"}, checkIO)
			if checkStatus != testCase.status {
				t.Errorf("check status = %d, want %d", checkStatus, testCase.status)
			}
			if got := checkIO.Stderr; got != "" {
				t.Errorf("check stderr = %q, want empty", got)
			}
			if got := checkIO.Stdout; got != testCase.checkWant {
				t.Errorf("check stdout = %q, want %q", got, testCase.checkWant)
			}

			indexIO := newLinkTestMem(testCase.body)
			indexStatus := Run([]string{"index", "d"}, indexIO)
			if indexStatus != 0 {
				t.Errorf("index status = %d, want 0", indexStatus)
			}
			if got := indexIO.Stderr; got != "" {
				t.Errorf("index stderr = %q, want empty", got)
			}
			if got := indexIO.Stdout; got != testCase.indexWant {
				t.Errorf("index stdout = %q, want %q", got, testCase.indexWant)
			}
		})
	}
}

func newLinkTestMem(body string) *MemIo {
	return NewMemIo(map[string]string{
		"d/s/SKILL.md":          "---\nname: s\ndescription: d\n---\n" + body + "\n",
		"d/s/ref.md":            "reference",
		"d/s/sub/f.md":          "nested",
		"d/s/percent%20name.md": "literal percent target",
		"d/s/%20":               "literal percent token",
	}, "d/s/special")
}
