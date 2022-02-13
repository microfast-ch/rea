package cpxtemplate

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"

	"github.com/davecgh/go-spew/spew"
)

type Tree struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Chardata string     `xml:",chardata"`
	Tree     []Tree     `xml:",any"`
}

func Parse(data []byte) {
	tree := &Tree{}

	err := xml.Unmarshal(data, tree)
	if err != nil {
		log.Fatalln(err)
	}

	spew.Dump(tree)
}

func summarize(tok xml.Token) string {
	switch tok.(type) {
	case xml.CharData:
		return "CharData: " + string([]byte(tok.(xml.CharData)))
	case xml.Comment:
		return "Comment: " + string([]byte(tok.(xml.Comment)))
	case xml.Directive:
		return "Directive: " + string([]byte(tok.(xml.Directive)))
	case xml.EndElement:
		return "EndElement: " + tok.(xml.EndElement).Name.Local
	case xml.ProcInst:
		return "ProcInst: " + tok.(xml.ProcInst).Target
	case xml.StartElement:
		return "StartElement: " + tok.(xml.StartElement).Name.Local
	default:
		return "unknown"
	}
}

type Node struct {
	// If the Token is xml.StartElement, it contains Nodes of certain tokens:
	// {CharData, Comment. etc} n-times and ends with xml.EndElement
	// e.g.
	//     Token: xml.StartElement
	//     Nodes: [xml.CharData, xml.Comment, xml.StartElement, xml.Comment, xml.EndElement]
	//             \Token: xml.CharData       \
	//                                        \Token: xml.StartElement
	//                                        \Nodes: [xml.CharData, xml.EndElement]
	Nodes  []*Node
	Token  xml.Token
	Parent *Node
}

func (n *Node) Append(tok xml.Token) *Node {
	newNode := &Node{Token: tok, Parent: n}
	n.Nodes = append(n.Nodes, newNode)
	return newNode
}

func (n *Node) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if n.Token != nil {
		if err := e.EncodeToken(n.Token); err != nil {
			if t, ok := n.Token.(xml.ProcInst); !ok {
				log.Println(t.Target)
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

func Parse2(data []byte) {
	d := xml.NewDecoder(bytes.NewReader(data))

	// On StartElement, add node to deepest node of the stack and append it itself to the stack.
	// On EndElement, add node to deepest node of the stack and pop one node from the stack
	// On others, add node to the deepest node of the stack
	// This should create a tree like this:
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

	root := &Node{}
	stack := []*Node{root}

	for {
		tokenInternal, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		// Internal bytes are only valid for the current scan
		tok := xml.CopyToken(tokenInternal)

		switch tok.(type) {
		case xml.StartElement:
			newChild := stack[len(stack)-1].Append(tok)
			stack = append(stack, newChild)
		case xml.EndElement:
			stack[len(stack)-1].Append(tok)
			stack = stack[:len(stack)-1]
		default:
			stack[len(stack)-1].Append(tok)
		}
	}
	spew.Dump(root)

	// Try to marshal it
	out, err := xml.Marshal(root)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(out))
}

func GetParent(child Node, parent string) *Node {
	// NIY
	return nil
}
