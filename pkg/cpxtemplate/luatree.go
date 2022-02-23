package cpxtemplate

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/djboris9/rea/pkg/xmltree"
)

// LuaTree represents an XML tree that is encoded into a lua program
type LuaTree struct {
	NodeList   []*xmltree.Node
	nodeListMx sync.Mutex

	LuaProg string // Lua program representing only the XML tree
}

// RegisterNode adds a xmltree node to the node registry of the lua tree,
// returning the new nodeId.
func (t *LuaTree) RegisterNode(node *xmltree.Node) uint32 {
	t.nodeListMx.Lock()
	t.NodeList = append(t.NodeList, node)
	nodeId := len(t.NodeList) - 1
	t.nodeListMx.Unlock()

	return uint32(nodeId)
}

// NewLuaTree converts an XML tree to a lua tree.
func NewLuaTree(tree *xmltree.Node) (*LuaTree, error) {
	lt := &LuaTree{}

	// Temporary lua script holder
	var sc strings.Builder

	// CharData tokenizer state
	tokenizerState := blockTokenizerCharBlock

	err := xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		// We register a node id for each node to keep track of it
		nodeId := lt.RegisterNode(node)

		// Indent script for better readability
		indent := strings.Repeat(" ", int(depth))

		switch v := node.Token.(type) {
		case xml.StartElement:
			fmt.Fprintf(&sc, "%sStartNode(%d) --  %v\n", indent, nodeId, v.Name.Local)
		case xml.EndElement:
			fmt.Fprintf(&sc, "%sEndNode(%d) --  %v\n", indent, nodeId, v.Name.Local)
		case xml.CharData:
			var err error
			tokenizerState, err = handleCharData(tokenizerState, &sc, indent, v, nodeId)
			if err != nil {
				return fmt.Errorf("processing node %d: %w", nodeId, err)
			}
		case xml.Directive, xml.Comment, xml.ProcInst:
			fmt.Fprintf(&sc, "%sSetToken(%d) -- Type: %T\n", indent, nodeId, v)
		default:
			return fmt.Errorf("unknown token type %T", v)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("converting XML Tree to LuaTree: %w", err)
	}

	lt.LuaProg = sc.String()

	return lt, nil
}

// BlockToken expresses a 2 char wide token that can be embedded inside CharData
type BlockToken string

const (
	BlockTokenStartCode  BlockToken = "[[" // Starts a code block
	BlockTokenEndCode    BlockToken = "]]" // Ends a code block
	BlockTokenStartPrint BlockToken = "[#" // Starts a printing block
	BlockTokenEndPrint   BlockToken = "#]" // Ends a printing block
)

// isToken checks if the given token is a non char data token.
func isToken(s string) bool {
	if len(s) != 2 {
		return false
	}

	tokens := []BlockToken{BlockTokenStartCode, BlockTokenEndCode, BlockTokenStartPrint, BlockTokenEndPrint}
	for i := range tokens {
		if s == string(tokens[i]) {
			return true
		}
	}

	return false
}

// codeBlockTokenizer splits the string d into strings with the tokens inside.
// Tokens are expected to be 2 chars long. The resulting slice contains at least one element.
// TODO: Add fuzzer
func codeBlockTokenizer(d string) []string {
	ret := []string{}

	lastToken := 0
	for idx := range d {
		if idx+1 >= len(d) {
			break
		}

		if lastToken > idx {
			// Skip round if we are still inside a token
			continue
		}

		curPos := d[idx : idx+2]
		if isToken(curPos) {
			ret = append(ret, d[lastToken:idx])
			ret = append(ret, curPos)
			lastToken = idx + 2
		}
	}

	ret = append(ret, d[lastToken:len(d)])

	return ret
}

// blockTokenizerState represents the context in which the block tokenizer is.
type blockTokenizerState int

const (
	blockTokenizerInvalid    blockTokenizerState = iota // We got an error
	blockTokenizerCharBlock  blockTokenizerState = iota // We are in a char data context
	blockTokenizerCodeBlock  blockTokenizerState = iota // We are in a code context
	blockTokenizerPrintBlock blockTokenizerState = iota // We are in a printing context
)

// handleCharData implements the block tokenizer that keeps track in which context
// the xml.CharData is and emits the according commands for the lua program.
func handleCharData(state blockTokenizerState, sc io.Writer, indent string, d xml.CharData, nodeId uint32) (blockTokenizerState, error) {
	toks := codeBlockTokenizer(string(d))

	// We got a CharData that has no blocks inside and are also in a chardata state.
	// Therefor we emit the token as is.
	if state == blockTokenizerCharBlock && len(toks) == 1 && !isToken(toks[0]) {
		fmt.Fprintf(sc, "%sSetToken(%d) --  %s\n", indent, nodeId, sanitizeComment(string(d)))
		return state, nil
	}

	for i := range toks {
		switch {
		case state == blockTokenizerCharBlock && !isToken(toks[i]):
			if toks[i] != "" {
				fmt.Fprintf(sc, "%sCharData(%d) --  %s\n", indent, nodeId, sanitizeComment(string(toks[i])))
			}
		case state == blockTokenizerCharBlock && toks[i] == string(BlockTokenStartCode):
			state = blockTokenizerCodeBlock
		case state == blockTokenizerCharBlock && toks[i] == string(BlockTokenStartPrint):
			state = blockTokenizerPrintBlock
		case state == blockTokenizerCodeBlock && toks[i] == string(BlockTokenEndCode):
			state = blockTokenizerCharBlock
		case state == blockTokenizerPrintBlock && toks[i] == string(BlockTokenEndPrint):
			state = blockTokenizerCharBlock
		case state == blockTokenizerCodeBlock && !isToken(toks[i]):
			fmt.Fprintf(sc, "%s%s -- CodeBlock\n", indent, toks[i])
		case state == blockTokenizerPrintBlock && !isToken(toks[i]):
			fmt.Fprintf(sc, "%sPrint(%s) -- PrintBlock\n", indent, toks[i])
		default:
			return blockTokenizerInvalid, fmt.Errorf("invalid token %q in tokenizer state %d", toks[i], state)
		}
	}

	return state, nil
}

// sanitizeComment returns a string that is suitable for a comment in the lua program.
func sanitizeComment(s string) string {
	if len(s) > 10 {
		s = "\"" + s[0:10] + "\"..."
	} else {
		s = "\"" + s + "\""
	}

	return strings.ReplaceAll(s, "\n", "\\n")
}
