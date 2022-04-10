package engine

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/djboris9/rea/pkg/xmltree"
)

// LuaTree represents an XML tree that is encoded into a lua program.
type LuaTree struct {
	NodeList   []*xmltree.Node
	nodeListMx sync.Mutex

	LuaProg string // Lua program representing only the XML tree
}

// RegisterNode adds a xmltree node to the node registry of the lua tree,
// returning the new nodeID.
func (t *LuaTree) RegisterNode(node *xmltree.Node) uint32 {
	t.nodeListMx.Lock()
	t.NodeList = append(t.NodeList, node)
	nodeID := len(t.NodeList) - 1
	t.nodeListMx.Unlock()

	return uint32(nodeID)
}

// NewLuaTree converts an XML tree to a lua tree.
func NewLuaTree(tree *xmltree.Node) (*LuaTree, error) {
	lt := &LuaTree{}

	// Temporary lua script holder
	var sc strings.Builder

	// Initialize FSM
	fsm := newFSM(&sc, lt.RegisterNode)

	err := xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		// We register a node id for each node to keep track of it
		nodeID := lt.RegisterNode(node)

		// Run an FSM step
		err := fsm.Next(nodeID, node, depth)
		if err != nil {
			return fmt.Errorf("executing FSM for node %d: %w", nodeID, err)
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

type nodeRegisterer func(*xmltree.Node) uint32

// Finite State Machine.
type luatreeFSM struct {
	inhibition []string
	state      luatreeFSMState
	sc         io.Writer
	curIndent  string
	registerer nodeRegisterer
}

func newFSM(buf io.Writer, registerer nodeRegisterer) *luatreeFSM {
	return &luatreeFSM{
		inhibition: []string{},
		state:      luatreeFSMStateChar,
		sc:         buf,
		registerer: registerer,
	}
}

func (fsm *luatreeFSM) Next(nodeID uint32, node *xmltree.Node, depth uint) error {
	fsm.curIndent = strings.Repeat(" ", int(depth))

	var err error

	switch v := node.Token.(type) {
	case xml.CharData:
		toks := codeBlockTokenizer(string(v))
		for i := range toks {
			switch toks[i] {
			case "":
				continue // Skip empty token processing
			case string(BlockTokenStartCode):
				err = fsm.processStartCode(nodeID, node)
			case string(BlockTokenEndCode):
				err = fsm.processEndCode(nodeID, node)
			case string(BlockTokenStartPrint):
				err = fsm.processStartPrint(nodeID, node)
			case string(BlockTokenEndPrint):
				err = fsm.processEndPrint(nodeID, node)
			default:
				err = fsm.processChar(nodeID, node, toks[i], len(toks) == 1)
			}

			if err != nil {
				return fmt.Errorf("state transition failed for token %q in node %d: %w", toks[i], nodeID, err)
			}
		}
	case xml.StartElement:
		err = fsm.processStartElement(nodeID, node)
	case xml.EndElement:
		err = fsm.processEndElement(nodeID, node)
	case xml.Directive, xml.Comment, xml.ProcInst:
		err = fsm.processNonstructuringElement(nodeID, node)
	default:
		err = fmt.Errorf("unknown token %T in node %d", v, nodeID)
	}

	if err != nil {
		return fmt.Errorf("state transition failed for node %d: %w", nodeID, err)
	}

	return nil
}

func (fsm *luatreeFSM) processStartElement(nodeID uint32, node *xmltree.Node) error {
	element, ok := node.Token.(xml.StartElement)
	if !ok {
		return errors.New("supplied wrong type for node")
	}

	tag := fmt.Sprintf("%sStartNode(%d) --  %v", fsm.curIndent, nodeID, element.Name.Local)

	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		// Inhibit elements if we are in a code or print state
		fsm.inhibition = append(fsm.inhibition, fmt.Sprintf("%s (inhibited)\n", tag))
	case luatreeFSMStateChar:
		fmt.Fprintf(fsm.sc, "%s\n", tag)
	}

	return nil
}

func (fsm *luatreeFSM) processEndElement(nodeID uint32, node *xmltree.Node) error {
	element, ok := node.Token.(xml.EndElement)
	if !ok {
		return errors.New("supplied wrong type for node")
	}

	tag := fmt.Sprintf("%sEndNode(%d) --  %v", fsm.curIndent, nodeID, element.Name.Local)

	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		// Inhibit elements if we are in a code or print state
		fsm.inhibition = append(fsm.inhibition, fmt.Sprintf("%s (inhibited)\n", tag))
	case luatreeFSMStateChar:
		fmt.Fprintf(fsm.sc, "%s\n", tag)
	}

	return nil
}

func (fsm *luatreeFSM) processChar(nodeID uint32, node *xmltree.Node, curToken string, singleToken bool) error {
	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		// If we are in a code or print context, we print the parts directly
		// as the envelope is handled in the according start and end blocks
		fmt.Fprintf(fsm.sc, "%s", curToken)
	case luatreeFSMStateChar:
		if singleToken {
			// Use `SetToken` instead of `CharData` if we have are in a xml.CharData without other tokens
			fmt.Fprintf(fsm.sc, "%sSetToken(%d) --  %s\n", fsm.curIndent, nodeID, sanitizeComment(curToken))
		} else {
			// Add new CharData node according to original one and set the data to our token content
			newNode := &xmltree.Node{
				Token:  xml.CharData(curToken),
				Parent: node.Parent,
			}
			newNodeID := fsm.registerer(newNode)
			fmt.Fprintf(fsm.sc, "%sCharData(%d) --  %s\n", fsm.curIndent, newNodeID, sanitizeComment(curToken))
		}
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processStartCode(nodeID uint32, node *xmltree.Node) error {
	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		return errors.New("start code block reached from inside a print or code block")
	case luatreeFSMStateChar:
		fmt.Fprintf(fsm.sc, "%s", fsm.curIndent)
		fsm.state = luatreeFSMStateCode
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processEndCode(nodeID uint32, node *xmltree.Node) error {
	switch fsm.state {
	case luatreeFSMStateChar, luatreeFSMStatePrint:
		return errors.New("end code block reached outside a code block")
	case luatreeFSMStateCode:
		fmt.Fprintf(fsm.sc, " -- CodeBlock\n")
		fsm.printInhibition()
		fsm.state = luatreeFSMStateChar
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processStartPrint(nodeID uint32, node *xmltree.Node) error {
	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		return errors.New("start print block reached from inside a print or code block")
	case luatreeFSMStateChar:
		fmt.Fprintf(fsm.sc, "%sPrint(", fsm.curIndent)
		fsm.state = luatreeFSMStatePrint
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processEndPrint(nodeID uint32, node *xmltree.Node) error {
	switch fsm.state {
	case luatreeFSMStateChar, luatreeFSMStateCode:
		return errors.New("end print block reached outside a code block")
	case luatreeFSMStatePrint:
		fmt.Fprintf(fsm.sc, ") -- PrintBlock\n")
		fsm.printInhibition()
		fsm.state = luatreeFSMStateChar
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processNonstructuringElement(nodeID uint32, node *xmltree.Node) error {
	tag := fmt.Sprintf("%sSetToken(%d) -- Type: %T", fsm.curIndent, nodeID, node.Token)

	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		fsm.inhibition = append(fsm.inhibition, fmt.Sprintf("%s (inhibited)\n", tag))
	case luatreeFSMStateChar:
		fmt.Fprintf(fsm.sc, "%s\n", tag)
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) printInhibition() {
	if len(fsm.inhibition) != 0 {
		fmt.Fprintf(fsm.sc, "%s", strings.Join(fsm.inhibition, ""))
		fsm.inhibition = []string{}
	}
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
