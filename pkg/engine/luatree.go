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

	// Initialize FSM
	fsm := newFSM(&sc, lt.RegisterNode)

	err := xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		// We register a node id for each node to keep track of it
		nodeId := lt.RegisterNode(node)

		// Run an FSM step
		err := fsm.Next(nodeId, node, depth)
		if err != nil {
			return fmt.Errorf("executing FSM for node %d: %w", nodeId, err)
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

// Finite State Machine
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

func (fsm *luatreeFSM) Next(nodeId uint32, node *xmltree.Node, depth uint) error {
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
				err = fsm.processStartCode(nodeId, node)
			case string(BlockTokenEndCode):
				err = fsm.processEndCode(nodeId, node)
			case string(BlockTokenStartPrint):
				err = fsm.processStartPrint(nodeId, node)
			case string(BlockTokenEndPrint):
				err = fsm.processEndPrint(nodeId, node)
			default:
				err = fsm.processChar(nodeId, node, toks[i], len(toks) == 1)
			}
			if err != nil {
				return fmt.Errorf("state transition failed for token %q in node %d: %w", toks[i], nodeId, err)
			}
		}
	case xml.StartElement:
		err = fsm.processStartElement(nodeId, node)
	case xml.EndElement:
		err = fsm.processEndElement(nodeId, node)
	case xml.Directive, xml.Comment, xml.ProcInst:
		err = fsm.processNonstructuringElement(nodeId, node)
	default:
		err = fmt.Errorf("unknown token %T in node %d", v, nodeId)
	}

	if err != nil {
		return fmt.Errorf("state transition failed for node %d: %w", nodeId, err)
	}

	return nil
}

func (fsm *luatreeFSM) processStartElement(nodeId uint32, node *xmltree.Node) error {
	element, ok := node.Token.(xml.StartElement)
	if !ok {
		return errors.New("supplied wrong type for node")
	}

	fmt.Fprintf(fsm.sc, "%sStartNode(%d) --  %v\n", fsm.curIndent, nodeId, element.Name.Local)
	// TODO: Inhibition
	return nil
}

func (fsm *luatreeFSM) processEndElement(nodeId uint32, node *xmltree.Node) error {
	element, ok := node.Token.(xml.EndElement)
	if !ok {
		return errors.New("supplied wrong type for node")
	}

	fmt.Fprintf(fsm.sc, "%sEndNode(%d) --  %v\n", fsm.curIndent, nodeId, element.Name.Local)
	// TODO: Inhibition
	return nil
}

func (fsm *luatreeFSM) processChar(nodeId uint32, node *xmltree.Node, curToken string, singleToken bool) error {
	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		// If we are in a code or print context, we print the parts directly
		// as the envelope is handled in the according start and end blocks
		fmt.Fprintf(fsm.sc, "%s", curToken)
	//case luatreeFSMStatePrint:
	// If we are in a print context, we print the code directly
	//		fmt.Fprintf(fsm.sc, "%sPrint(%s) -- PrintBlock\n", fsm.curIndent, curToken)
	case luatreeFSMStateChar:
		if singleToken {
			// Use `SetToken` instead of `CharData` if we have are in a xml.CharData without other tokens
			fmt.Fprintf(fsm.sc, "%sSetToken(%d) --  %s\n", fsm.curIndent, nodeId, sanitizeComment(curToken))
		} else {
			// Add new CharData node according to original one and set the data to our token content
			newNode := &xmltree.Node{
				Token:  xml.CharData(curToken),
				Parent: node.Parent,
			}
			newNodeId := fsm.registerer(newNode)
			fmt.Fprintf(fsm.sc, "%sCharData(%d) --  %s\n", fsm.curIndent, newNodeId, sanitizeComment(curToken))
		}
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processStartCode(nodeId uint32, node *xmltree.Node) error {
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

func (fsm *luatreeFSM) processEndCode(nodeId uint32, node *xmltree.Node) error {
	switch fsm.state {
	case luatreeFSMStateChar, luatreeFSMStatePrint:
		return errors.New("end code block reached outside a code block")
	case luatreeFSMStateCode:
		fmt.Fprintf(fsm.sc, " -- CodeBlock\n")
		fsm.state = luatreeFSMStateChar
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processStartPrint(nodeId uint32, node *xmltree.Node) error {
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

func (fsm *luatreeFSM) processEndPrint(nodeId uint32, node *xmltree.Node) error {
	switch fsm.state {
	case luatreeFSMStateChar, luatreeFSMStateCode:
		return errors.New("end print block reached outside a code block")
	case luatreeFSMStatePrint:
		fmt.Fprintf(fsm.sc, ") -- PrintBlock\n")
		fsm.state = luatreeFSMStateChar
	default:
		return errors.New("invalid state")
	}

	return nil
}

func (fsm *luatreeFSM) processNonstructuringElement(nodeId uint32, node *xmltree.Node) error {
	switch fsm.state {
	case luatreeFSMStateCode, luatreeFSMStatePrint:
		return errors.New("special element block reached inside a code or print block") // TODO: Do inhibition!
	case luatreeFSMStateChar:
		fmt.Fprintf(fsm.sc, "%sSetToken(%d) -- Type: %T\n", fsm.curIndent, nodeId, node.Token)
	default:
		return errors.New("invalid state")
	}
	return nil
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
