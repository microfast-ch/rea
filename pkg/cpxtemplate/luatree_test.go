package cpxtemplate

import (
	"encoding/xml"
	"testing"

	"github.com/djboris9/rea/pkg/xmltree"
)

func TestNewLuaTree(t *testing.T) {
	testdata := xml.Header + `
<p1>
  <p2 no="1">Inside P2</p2>
  <p2 no="2">Inside P2 again</p2>
  <p2 no="3"><p3>Inside P3</p3></p2>
  <p2 no="4" be="5">Before P3 <p3>Inside P3</p3> after P3</p2>
  <!-- my comment :) -->
</p1>`

	tree, err := xmltree.Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	NewLuaTree(tree)
}
