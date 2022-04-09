package engine

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

	// Collector for inhibited calls
	inhibitor := ""

	err := xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		// We register a node id for each node to keep track of it
		nodeId := lt.RegisterNode(node)

		// Indent script for better readability
		indent := strings.Repeat(" ", int(depth))

		// Ignore tokens that slipped inside of a code or print block
		if _, ok := node.Token.(xml.CharData); (tokenizerState == blockTokenizerCodeBlock || tokenizerState == blockTokenizerPrintBlock) && !ok {
			// TODO: Collect inhibited calls to StartNode, EndNode, SetToken
			inhibitor += fmt.Sprintf("%T,", node.Token)
			return nil
		}

		// We had something collected in the inhibitor and now we are out of the code or print block.
		// So print the inhibited calls and reset the inhibitor
		// TODO: We should print the collected inhibited calls after we left the print or code block
		if inhibitor != "" && tokenizerState == blockTokenizerCharBlock {
			fmt.Fprintf(&sc, "%s -- Inhibited calls to: %s\n", indent, inhibitor)
			inhibitor = ""
		}

		switch v := node.Token.(type) {
		case xml.StartElement:
			fmt.Fprintf(&sc, "%sStartNode(%d) --  %v\n", indent, nodeId, v.Name.Local)
		case xml.EndElement:
			fmt.Fprintf(&sc, "%sEndNode(%d) --  %v\n", indent, nodeId, v.Name.Local)
		case xml.CharData:
			var err error
			tokenizerState, err = handleCharData(lt, tokenizerState, &sc, indent, nodeId, node)
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

type luatreeFSMState int

const (
	luatreeFSMStateChar  luatreeFSMState = iota // We are directly printing chars blocks, this is the default state
	luatreeFSMStateCode  luatreeFSMState = iota // We are in a code block
	luatreeFSMStatePrint luatreeFSMState = iota // We are in a print block
)

// Finite State Machine
type luatreeFSM struct {
	inhibition []string
	state      luatreeFSMState
	sc         *io.Writer
	indent     string
}

func (fsm *luatreeFSM) Run() error {
	for {
		token := *xmltree.Node{} // TODO: Feed this from walk function

		var err error
		switch v := node.Token.(type) {
		case xml.CharData:
			toks := codeBlockTokenizer(string(v))
			for i := range toks {
				switch toks[i] {
				case BlockTokenStartCode:
					err = fsm.processStartCode()
				case BlockTokenEndCode:
					err = fsm.processEndCode()
				case BlockTokenStartPrint:
					err = fsm.processStartCode()
				case BlockTokenEndPrint:
					err = fsm.processEndCode()
				default:
					err = fsm.processChar(toks[i], len(toks) == 1)
				}
				if err != nil {
					return fmt.Errorf("state transition: %w", err) // TODO message
				}
			}
		case xml.StartElement:
			err = fsm.processStartElement(v)
		case xml.EndElement:
			err = fsm.processEndElement(v)
		case xml.Directive, xml.Comment, xml.ProcInst:
			err = fsm.processNonstructuringElement(v)
		default:
			err = fmt.Errorf("unknown token %T", v)
		}

		if err != nil {
			return fmt.Errorf("state transition: %w", err) // TODO message
		}

	}
}

func (fsm *luatreeFSM) processChar(nodeId uint32, node *xmltree.Node, curToken []byte, singleToken bool) error {
	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		// If we are in a code or print context, we must directly
		// print the value
		fmt.Fprintf(fsm.sc, "%s", next)
	case luatreeFSMStateChar:
		if singleToken {
			fmt.Fprintf(&sc, "%sSetToken(%d) -- Type: %T\n", fsm.indent, nodeId, node)
		} else {
			// Add new CharData node according to original one and set the data to toks[i]
			newNode := &xmltree.Node{
				Token:  xml.CharData(curToken),
				Parent: node.Parent,
			}
			newNodeId := lt.RegisterNode(newNode)
			fmt.Fprintf(sc, "%sCharData(%d) --  %s\n", fsm.indent, newNodeId, sanitizeComment(curToken))
		}
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processStartCode() error {
	// TODO: Check we came from a valid state

	fsm.state = luatreeFSMStateCode

	return nil
}

func (fsm *luatreeFSM) processNonstructuringElement(nodeId uint32, node *xmltree.Node) error {
	// TODO: Check we came from a valid state
	fmt.Fprintf(&fsm.sc, "%sSetToken(%d) -- Type: %T\n", fsm.indent, nodeId, node)
	return nil
}

// handleCharData implements the block tokenizer that keeps track in which context
// the xml.CharData is and emits the according commands for the lua program.
func handleCharData(lt *LuaTree, state blockTokenizerState, sc io.Writer, indent string, nodeId uint32, node *xmltree.Node) (blockTokenizerState, error) {
	d := node.Token.(xml.CharData)

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
			// Add new CharData node according to original one and set the data to toks[i]
			if toks[i] != "" {
				newNode := &xmltree.Node{
					Token:  xml.CharData(toks[i]),
					Parent: node.Parent,
				}
				newNodeId := lt.RegisterNode(newNode)
				fmt.Fprintf(sc, "%sCharData(%d) --  %s\n", indent, newNodeId, sanitizeComment(string(toks[i])))
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
