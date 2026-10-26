package main

import "testing"

func wantShown(t *testing.T, input, want string) {
	t.Helper()
	if got := shown(input); got != want {
		t.Errorf("shown(%q) = %q, want %q", input, got, want)
	}
}

func TestShownReplacesControls(t *testing.T) {
	wantShown(t, "a\x1bb", "a?b")
	wantShown(t, "a\x7fb", "a?b")
	wantShown(t, "a"+string(rune(0x0080))+"b", "a?b")
	wantShown(t, "a"+string(rune(0x0085))+"b", "a?b")
	wantShown(t, "a"+string(rune(0x009f))+"b", "a?b")
}

func TestShownReplacesSeparatorsAndBidiControls(t *testing.T) {
	wantShown(t, "a"+string(rune(0x2028))+"b", "a?b")
	wantShown(t, "a"+string(rune(0x2029))+"b", "a?b")
	wantShown(t, "a"+string(rune(0x202a))+"b", "a?b")
	wantShown(t, "a"+string(rune(0x202e))+"b", "a?b")
	wantShown(t, "a"+string(rune(0x2066))+"b", "a?b")
	wantShown(t, "a"+string(rune(0x2069))+"b", "a?b")
}

func TestShownKeepsNeighbouringCharacters(t *testing.T) {
	wantShown(t, "caf"+string(rune(0x00e9))+" "+string(rune(0x1F600)), "caf"+string(rune(0x00e9))+" "+string(rune(0x1F600)))
	wantShown(t, "a"+string(rune(0x00a0))+"b", "a"+string(rune(0x00a0))+"b")
	wantShown(t, "a"+string(rune(0x200b))+"b", "a"+string(rune(0x200b))+"b")
	wantShown(t, "a"+string(rune(0x200e))+"b", "a"+string(rune(0x200e))+"b")
	wantShown(t, "a"+string(rune(0x202f))+"b", "a"+string(rune(0x202f))+"b")
	wantShown(t, "a"+string(rune(0x2065))+"b", "a"+string(rune(0x2065))+"b")
	wantShown(t, "a"+string(rune(0x206a))+"b", "a"+string(rune(0x206a))+"b")
}

func TestTableKeepsColumnsAlignedAroundHazards(t *testing.T) {
	got := table([]string{"A", "B"}, [][]string{{"x" + string(rune(0x202e)) + "y", "1"}, {"zz", "2"}})
	want := "A    B\nx?y  1\nzz   2\n"
	if got != want {
		t.Errorf("table = %q, want %q", got, want)
	}
}
