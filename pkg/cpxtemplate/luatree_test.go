package cpxtemplate

import (
	"encoding/xml"
	"testing"

	"github.com/djboris9/rea/pkg/xmltree"
	"github.com/google/go-cmp/cmp"
)

func TestNewLuaTree(t *testing.T) {
	testdata := xml.Header + `
<p1>
  <p2 no="1">Inside P2</p2>
  <p2 no="2">Inside P2 again</p2>
  <p2 no="3"><p3>Inside P3</p3></p2>
  <p2 no="4" be="5">Before P3 <p3>Inside P3</p3> after P3</p2>
  <!-- my comment :) -->
  <p2 no="5">[[ if (A) ]]Hallo [# A #]</p2>
  <p2 no="6">[[ endif ]]</p2>
</p1>`

	tree, err := xmltree.Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	lt, err := NewLuaTree(tree)
	if err != nil {
		t.Error(err)
	}

	t.Log(lt.LuaProg)
	// TODO: Validate output of LuaProg and the node list
}

func TestCodeBlockTokenizer(t *testing.T) {
	got := codeBlockTokenizer("abcd[[ efg ]]hi[# jk #]lmn")
	want := []string{"abcd", "[[", " efg ", "]]", "hi", "[#", " jk ", "#]", "lmn"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("codeBlockTokenizer() mismatch (-want +got):\n%s", diff)
	}

	got = codeBlockTokenizer("")
	want = []string{""}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("codeBlockTokenizer() mismatch (-want +got):\n%s", diff)
	}

	got = codeBlockTokenizer("hello")
	want = []string{"hello"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("codeBlockTokenizer() mismatch (-want +got):\n%s", diff)
	}
}
