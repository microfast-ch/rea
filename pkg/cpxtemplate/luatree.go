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

// BlockToken expresses a 2 char wide token that can be embedded inside CharData
type BlockToken string

const (
	BlockTokenStartCode  = "[["
	BlockTokenEndCode    = "]]"
	BlockTokenStartPrint = "[#"
	BlockTokenEndPrint   = "#]"
)

func isToken(s string) bool {
	if len(s) != 2 {
		return false
	}

	tokens := []string{BlockTokenStartCode, BlockTokenEndCode, BlockTokenStartPrint, BlockTokenEndPrint}
	for i := range tokens {
		if s == tokens[i] {
			return true
		}
	}

	return false
}

// codeBlockTokenizer splits the string d into strings with the tokens inside.
// Tokens are expected to be 2 chars long. The resulting slice contains at least one element.
func codeBlockTokenizer(d string) []string {
	tokens := []string{BlockTokenStartCode, BlockTokenEndCode, BlockTokenStartPrint, BlockTokenEndPrint}

	ret := []string{}

	lastToken := 0
	for idx := range d {
		if idx+1 >= len(d) {
			break
		}

		curPos := d[idx : idx+2]
		for tok := range tokens {
			if tokens[tok] == curPos {
				ret = append(ret, d[lastToken:idx])
				ret = append(ret, tokens[tok])
				lastToken = idx + 2
				break
			}
		}
	}

	ret = append(ret, d[lastToken:len(d)])

	return ret
}

type blockTokenizerState int

const (
	blockTokenizerInvalid    blockTokenizerState = iota
	blockTokenizerCharBlock  blockTokenizerState = iota
	blockTokenizerCodeBlock  blockTokenizerState = iota
	blockTokenizerPrintBlock blockTokenizerState = iota
)

// This can only handle tokens that start and end in the same block
func handleCharData(state blockTokenizerState, sc io.Writer, indent string, d xml.CharData, nodeId uint32) (blockTokenizerState, error) {
	toks := codeBlockTokenizer(string(d))

	// We got a CharData that has no blocks inside and are also in a chardata state.
	// Therefor we emit the token as is.
	if state == blockTokenizerCharBlock && len(toks) == 1 && !isToken(toks[0]) {
		fmt.Fprintf(sc, "%sToken(%d) --  %s\n", indent, nodeId, sanitizeComment(string(d)))
		return state, nil
	}

	for i := range toks {
		switch {
		case state == blockTokenizerCharBlock && !isToken(toks[i]):
			if toks[i] != "" {
				fmt.Fprintf(sc, "%sCharData(%d) --  %s\n", indent, nodeId, sanitizeComment(string(toks[i])))
			}
		case state == blockTokenizerCharBlock && toks[i] == BlockTokenStartCode:
			state = blockTokenizerCodeBlock
		case state == blockTokenizerCharBlock && toks[i] == BlockTokenStartPrint:
			state = blockTokenizerPrintBlock
		case state == blockTokenizerCodeBlock && toks[i] == BlockTokenEndCode:
			state = blockTokenizerCharBlock
		case state == blockTokenizerPrintBlock && toks[i] == BlockTokenEndPrint:
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

func NewLuaTree(tree *xmltree.Node) error {
	lt := &LuaTree{}

	// Temporary lua script holder
	var sc strings.Builder

	// CharData tokenizer state
	tokenizerState := blockTokenizerCharBlock

	err := xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
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
			var err error
			tokenizerState, err = handleCharData(tokenizerState, &sc, indent, v, nodeId)
			if err != nil {
				return fmt.Errorf("processing node %d: %w", nodeId, err)
			}
		case xml.Directive, xml.Comment, xml.ProcInst:
			fmt.Fprintf(&sc, "%sToken(%d) -- Type: %T\n", indent, nodeId, v)
		default:
			return fmt.Errorf("unknown token type %T", v)
		}

		return nil
	})

	log.Printf("\n%s", sc.String())
	if err != nil {
		return fmt.Errorf("converting XML Tree to LuaTree: %w", err)
	}
	return nil
}
