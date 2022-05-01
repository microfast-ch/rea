package engine

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCodeBlockTokenizer(t *testing.T) {
	// Complex 1
	got := codeBlockTokenizer("abcd[[ efg ]]hi[# jk #]lmn")
	want := []string{"abcd", "[[", " efg ", "]]", "hi", "[#", " jk ", "#]", "lmn"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("codeBlockTokenizer() mismatch (-want +got):\n%s", diff)
	}

	// Empty
	got = codeBlockTokenizer("")
	want = []string{""}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("codeBlockTokenizer() mismatch (-want +got):\n%s", diff)
	}

	// Basic
	got = codeBlockTokenizer("hello")
	want = []string{"hello"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("codeBlockTokenizer() mismatch (-want +got):\n%s", diff)
	}

	// Complex 2
	// TODO: We might remove the starting and trailing "" from the result, but it doesn't hurt that much
	got = codeBlockTokenizer("[[ for i=1,3 do ]]X[# i #]Y[[ end ]]")
	want = []string{"", "[[", " for i=1,3 do ", "]]", "X", "[#", " i ", "#]", "Y", "[[", " end ", "]]", ""}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("codeBlockTokenizer() mismatch (-want +got):\n%s", diff)
	}
}
