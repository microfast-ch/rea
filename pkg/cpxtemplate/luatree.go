package cpxtemplate

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
	"sync/atomic"

	"github.com/djboris9/rea/pkg/xmltree"
)

type LuaTree struct {
	lastNodeId uint32
}

func (t *LuaTree) RegisterNode(node *xmltree.Node) uint32 {
	nodeId := atomic.AddUint32(&t.lastNodeId, 1)
	return nodeId
}

func sanitizeComment(s string) string {
	if len(s) > 10 {
		s = "\"" + s[0:10] + "\"..."
	} else {
		s = "\"" + s + "\""
	}

	return strings.ReplaceAll(s, "\n", "\\n")
}

// This can only handle tokens that start and end in the same block
func handleCharData(sc io.Writer, indent string, d xml.CharData, nodeId uint32) {
	// TODO:
	// - Print token when CharData is not in a code context and has none inside
	// - Start code context when token is machted
	// - Keep track of code context over multiple sequential CharDatas
	fmt.Fprintf(sc, "%sCharData(%d) --  %s\n", indent, nodeId, sanitizeComment(string(d)))
}

func NewLuaTree(tree *xmltree.Node) {
	lt := &LuaTree{}

	// Temporary lua script holder
	var sc strings.Builder

	xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		// We register a node id for each node to keep track of it
		nodeId := lt.RegisterNode(node)

		// indent script for better readability
		indent := strings.Repeat(" ", int(depth))

		switch v := node.Token.(type) {
		case xml.StartElement:
			fmt.Fprintf(&sc, "%sStartNode(%d) --  %v\n", indent, nodeId, v.Name.Local)
		case xml.EndElement:
			fmt.Fprintf(&sc, "%sEndNode(%d) --  %v\n", indent, nodeId, v.Name.Local)
		case xml.CharData:
			handleCharData(&sc, indent, v, nodeId)
		case xml.Directive, xml.Comment, xml.ProcInst:
			fmt.Fprintf(&sc, "%sToken(%d) -- Type: %T\n", indent, nodeId, v)
		default:
			return fmt.Errorf("unknown token type %T", v)
		}

		return nil
	})
	log.Printf("\n%s", sc.String())
}
