package xmltree

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"

	"github.com/davecgh/go-spew/spew"
)

// Node represents a XML tree, which can have subnodes and connets it to its
// parent node.
//
// If the Token is of type xml.StartElement, it can have Nodes (children) of
// other nodes with different types, but has to end with a node of type xml.EndElement.
// Example:
//     Token: xml.StartElement
//     Nodes: [xml.CharData, xml.Comment, xml.StartElement, xml.Comment, xml.EndElement]
//                                        \
//                                        \Token: xml.StartElement
//                                        \Nodes: [xml.CharData, xml.EndElement]
//
// This creates a tree like this:
//     CharData
//     Comment
//     StartNode A
//       CharData
//       StartNode B
//         CharData
//         EndNode B
//       CharData
//       EndNode A
//     CharData
type Node struct {
	Nodes  []*Node
	Token  xml.Token
	Parent *Node
}

// Append adds the given token to the node childrens list, returning the
// newly appended child.
func (n *Node) Append(tok xml.Token) *Node {
	newNode := &Node{Token: tok, Parent: n}
	n.Nodes = append(n.Nodes, newNode)

	return newNode
}

// MarshalXML implements the xml.Marshaler interface, so Node objects can be
// marshalled conveniently with `xml.Marshal(node)`.
// The output of marshal should result semantically to the originally parsed document.
func (n *Node) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if n.Token != nil {
		if err := e.EncodeToken(n.Token); err != nil {
			if t, ok := n.Token.(xml.ProcInst); !ok {
				log.Println(t.Target) // TODO: Why is this here?
			}

			return err
		}
	}

	for i := range n.Nodes {
		if err := n.Nodes[i].MarshalXML(e, start); err != nil {
			return err
		}
	}

	return nil
}

// Dump returns the node structure in a human readable format. The format can
// change between versions.
func (n *Node) Dump() string {
	return spew.Sdump(n)
}

// Copy deep copies the node, anchoring at the new given parent.
func (n *Node) Copy(parent *Node) *Node {
	res := &Node{
		Parent: n.Parent,
		Token:  xml.CopyToken(n.Token),
		Nodes:  make([]*Node, len(n.Nodes)),
	}

	for i := range n.Nodes {
		res.Nodes[i] = n.Nodes[i].Copy(res)
	}

	return res
}

// Parse translates the given xml document to a XML tree.
func Parse(data []byte) (*Node, error) {
	d := xml.NewDecoder(bytes.NewReader(data))

	// On StartElement, add node to deepest node of the stack and append it itself to the stack.
	// On EndElement, add node to deepest node of the stack and pop one node from the stack
	// On others, add node to the deepest node of the stack

	root := &Node{}
	stack := []*Node{root}

	for {
		tokenInternal, err := d.Token()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("reading xml token: %w", err)
		}

		// Internal bytes are only valid for the current scan
		tok := xml.CopyToken(tokenInternal)

		switch tok.(type) {
		case xml.StartElement:
			newChild := stack[len(stack)-1].Append(tok)
			stack = append(stack, newChild)
		case xml.EndElement:
			stack[len(stack)-1].Append(tok)
			stack = stack[:len(stack)-1] // TODO: Can we be out of range here?
		default:
			stack[len(stack)-1].Append(tok)
		}
	}

	return root, nil
}

// WalkFunc defines the function that can be called by a walk function on a xml tree.
type WalkFunc func(node *Node, depth uint) error

// Walk traverses deep-first the given xml tree. On each traversed node, it
// calls the WalkFunc. If the WalkFunc returns an error, the traversal stops
// and Walk returns the wrapped reported error back.
func Walk(root *Node, fn WalkFunc) error {
	return walk(root, fn, 0)
}

func walk(root *Node, fn WalkFunc, depth uint) error {
	err := fn(root, depth)
	if err != nil {
		return fmt.Errorf("error executing WalkFunc: %w", err)
	}

	for i := range root.Nodes {
		err := walk(root.Nodes[i], fn, depth+1)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetParent returns the first parent node of the given child, having
// the given localName.
func GetParent(child *Node, localName string) *Node {
	if child == nil {
		return nil
	}

	for curNode := child.Parent; curNode != nil; curNode = curNode.Parent {
		tok, ok := curNode.Token.(xml.StartElement)
		if ok && tok.Name.Local == localName {
			return curNode
		}
	}

	return nil
}
