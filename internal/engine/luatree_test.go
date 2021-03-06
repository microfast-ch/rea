package engine

import (
	"encoding/xml"
	"testing"

	"github.com/djboris9/xmltree"
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

	want := ` SetToken(1) -- Type: xml.ProcInst
 SetToken(2) --  "\n\n"
 StartNode(3) --  p1
  SetToken(4) --  "\n  "
  StartNode(5) --  p2
   SetToken(6) --  "Inside P2"
   EndNode(7) --  p2
  SetToken(8) --  "\n  "
  StartNode(9) --  p2
   SetToken(10) --  "Inside P2 "...
   EndNode(11) --  p2
  SetToken(12) --  "\n  "
  StartNode(13) --  p2
   StartNode(14) --  p3
    SetToken(15) --  "Inside P3"
    EndNode(16) --  p3
   EndNode(17) --  p2
  SetToken(18) --  "\n  "
  StartNode(19) --  p2
   SetToken(20) --  "Before P3 "
   StartNode(21) --  p3
    SetToken(22) --  "Inside P3"
    EndNode(23) --  p3
   SetToken(24) --  " after P3"
   EndNode(25) --  p2
  SetToken(26) --  "\n  "
  SetToken(27) -- Type: xml.Comment
  SetToken(28) --  "\n  "
  StartNode(29) --  p2
    if (A)  -- CodeBlock
   CharData(31) --  "Hallo "
   Print( A ) -- PrintBlock
   EndNode(32) --  p2
  SetToken(33) --  "\n  "
  StartNode(34) --  p2
    endif  -- CodeBlock
   EndNode(36) --  p2
  SetToken(37) --  "\n"
  EndNode(38) --  p1
`

	tree, err := xmltree.Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	lt, err := NewLuaTree(tree)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(want, lt.LuaProg); diff != "" {
		t.Errorf("generated LuaProg mismatch (-want +got):\n%s", diff)
	}
}
