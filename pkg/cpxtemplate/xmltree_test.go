package cpxtemplate

import (
	"encoding/xml"
	"github.com/google/go-cmp/cmp"
	"testing"
)

const testdata = xml.Header + `
<p1>
  <p2 no="1">Inside P2</p2>
  <p2 no="2">Inside P2 again</p2>
  <p2 no="3"><p3>Inside P3</p3></p2>
  <p2 no="4" be="5">Before P3 <p3>Inside P3</p3> after P3</p2>
  <!-- my comment :) -->
</p1>`

func TestParse(t *testing.T) {
	tree, err := Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	got, err := xml.Marshal(tree)
	if err != nil {
		t.Fatalf("marshaling tree: %v", err)
	}

	if diff := cmp.Diff(string(testdata), string(got)); diff != "" {
		t.Errorf("Parse() and marshaling mismatch (-want +got):\n%s", diff)
	}
}

func TestWalk(t *testing.T) {
	tree, err := Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	count := 0
	fn := func(n *Node, depth uint) error {
		if n != nil {
			count++
		}

		return nil
	}

	err = Walk(tree, fn)
	if err != nil {
		t.Fatalf("walk got error: %v", err)
	}

	if count != 30 {
		t.Errorf("should visited 30 nodes, but got %d", count)
	}
}

func TestGetParent(t *testing.T) {
	tree, err := Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	// Get a known child and parent node
	// |-parent---------|              |-child--|
	// <p2 no="4" be="5">Before P3 <p3>Inside P3</p3> after P3</p2>
	var child *Node
	var parent *Node
	fn := func(n *Node, depth uint) error {
		if tok, ok := n.Token.(xml.StartElement); ok {
			if tok.Name.Local == "p2" && len(tok.Attr) == 2 && tok.Attr[1].Name.Local == "be" {
				parent = n
			}
		}

		if tok, ok := n.Token.(xml.CharData); ok {
			if string(tok) == "Inside P3" {
				child = n
			}
		}

		return nil
	}

	err = Walk(tree, fn)
	if err != nil {
		t.Fatalf("walk got error: %v", err)
	}

	if child == nil || parent == nil {
		t.Fatal("cannot determine initial parent and child")
	}

	gotParent := GetParent(child, "p2")
	if gotParent != parent {
		t.Errorf("Wanted node %p, got %p as parent of %p", parent, gotParent, child)
	}

}

func TestNodeModification(t *testing.T) {
	var want = xml.Header + `
<p1>
  <p2 no="1">Inside P2</p2>
  <p2 no="2">Inside P2 again</p2>
  <p2 no="3"><p3>Inside P3</p3></p2>
  <p2 no="4" be="5">Before P3 <p3>Inside P3</p3> after P3</p2><p3>copied outside</p3>
  <!-- my comment :) -->
</p1>`

	// Parse testdata
	tree, err := Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	// Get a known child
	//                             |----child-------|
	// <p2 no="4" be="5">Before P3 <p3>Inside P3</p3> after P3</p2>
	var child *Node
	fn := func(n *Node, depth uint) error {
		if tok, ok := n.Token.(xml.StartElement); ok {
			if tok.Name.Local == "p3" {
				child = n
			}
		}

		return nil
	}

	err = Walk(tree, fn)
	if err != nil {
		t.Fatalf("walk got error: %v", err)
	}

	if child == nil {
		t.Fatal("cannot determine initial child")
	}

	// Modify structure and chardata to check if the object is not copied by reference
	newChild := child.Copy(child.Parent)
	tok := newChild.Nodes[0].Token.(xml.CharData)
	tok = []byte("copied outside")
	newChild.Nodes[0].Token = tok

	child.Parent.Nodes = append(child.Parent.Nodes, newChild)

	// Marshal data
	got, err := xml.Marshal(tree)
	if err != nil {
		t.Fatalf("marshaling tree: %v", err)
	}

	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("Parse() and marshaling mismatch (-want +got):\n%s", diff)
	}
}
