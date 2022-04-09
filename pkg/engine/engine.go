package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/Shopify/go-lua"
	"github.com/djboris9/rea/pkg/xmltree"
)

type LuaEngine struct {
	lt       *LuaTree
	luaState *lua.State

	// State information during exec
	execLock    sync.Mutex
	nodePath    []*xmltree.Node
	parentStack []*xmltree.Node // stack of StartNode elements to keep the balance
	nodePathStr []string
}

func NewLuaEngine(lt *LuaTree) *LuaEngine {
	// Initialize lua
	l := lua.NewState()
	//lua.OpenLibraries(l)
	//lua.BaseOpen(l)

	e := &LuaEngine{
		lt:       lt,
		luaState: l,
	}

	// Map our Go functions to lua
	l.Register("SetToken", e.iSetToken)
	l.Register("StartNode", e.iStartNode)
	l.Register("EndNode", e.iEndNode)
	l.Register("CharData", e.iCharData)
	l.Register("Print", e.iPrint)

	// Restricted base library
	l.Register("tostring", baseToString) // Required for Print

	// Return engine
	return e
}

// Extract from https://github.com/Shopify/go-lua/blob/9ab7793778076a5d7bd05bae27462473a0a29a4a/base.go#L301
func baseToString(l *lua.State) int {
	lua.CheckAny(l, 1)
	lua.ToStringMeta(l, 1)
	return 1
}

// This function is serialized on exec
func (e *LuaEngine) Exec() error {
	e.execLock.Lock() // TODO: We might convert it to sync.Once
	defer e.execLock.Unlock()

	// Initialize empty state
	e.nodePath = []*xmltree.Node{}
	e.nodePathStr = []string{}

	// Execute lua program
	err := lua.DoString(e.luaState, e.lt.LuaProg)
	if err != nil {
		// We got an error, but the detailed error message is on the stack. Wrap it.
		return fmt.Errorf("executing lua prog got %w with :%s", err, lua.CheckString(e.luaState, -1))
	}

	return err
}

// Can only be called after Exec() has been run
func (e *LuaEngine) WriteXML(w io.Writer) error {
	enc := xml.NewEncoder(w)
	for i := range e.nodePath {
		//if err := EncodeToken(enc, w, e.nodePath[i].Token); err != nil { // Is a custom encoder needed
		if err := enc.EncodeToken(e.nodePath[i].Token); err != nil {
			return fmt.Errorf("encoding token %d: %w", i, err)
		}
	}

	err := enc.Flush()
	if err != nil {
		return fmt.Errorf("flushing encoder: %w", err)
	}

	return nil
}

// Can only be called after Exec() has been run
func (e *LuaEngine) GetNodePath() []*xmltree.Node {
	return e.nodePath
}

// Can only be called after Exec() has been run
func (e *LuaEngine) GetNodePathString() []string {
	return e.nodePathStr
}

func (e *LuaEngine) iStartNode(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeId]

	// Fill tree, before we add us to the parent stack
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("StartNode(%d)", nodeId))

	// Add node to the parent stack
	e.parentStack = append(e.parentStack, node)

	return 0
}

func (e *LuaEngine) iEndNode(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeId]

	// Fill tree, before we remove the last parent from the stack
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("EndNode(%d)", nodeId))

	// Remove one level from the parent stack
	e.parentStack = e.parentStack[:len(e.parentStack)-1]

	return 0
}

func (e *LuaEngine) iSetToken(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeId]

	// Fill tree, before we are added to the path
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("SetToken(%d)", nodeId))

	return 0
}

func (e *LuaEngine) iCharData(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeId]

	// Fill tree, before we are added to the path
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("CharData(%d)", nodeId))

	return 0
}

func (e *LuaEngine) iPrint(state *lua.State) int {
	var sc strings.Builder

	// Based on https://github.com/Shopify/go-lua/blob/9ab7793778076a5d7bd05bae27462473a0a29a4a/base.go#L205
	n := state.Top()
	state.Global("tostring")
	for i := 1; i <= n; i++ {
		state.PushValue(-1) // function to be called
		state.PushValue(i)  // value to print
		state.Call(1, 1)
		s, ok := state.ToString(-1)
		if !ok {
			lua.Errorf(state, "'tostring' must return a string to 'print'")
			panic("unreachable")
		}
		if i > 1 {
			sc.WriteString("\t")
		}
		sc.WriteString(s)
		state.Pop(1) // pop result
	}

	e.nodePathStr = append(e.nodePathStr, "Print(???)")

	// Create new node and add it to the nodePath
	chrNode := xml.CharData(sc.String())
	node := &xmltree.Node{
		Token:  chrNode,
		Parent: e.parentStack[len(e.parentStack)-1],
	}
	e.nodePath = append(e.nodePath, node)

	return 0
}

// On each node this function is called. It checks if the new node has the same
// parent as the previous one. If not, it determines the common parent and closes
// the tags parent tags up to it on the old branch. Then it opens all start tags of the new
// branch so the new tag can be added.
// This needs to be called before the node is added to the path.
// TODO: Handle StartNode and EndNodes correctly for the tree
func (e *LuaEngine) fillTree(newNode *xmltree.Node) {
	// We are the root or are somehow detached. No balancing possible.
	if newNode.Parent == nil || len(e.parentStack) == 0 {
		return
	}

	previousNode := e.nodePath[len(e.nodePath)-1]
	lastStack := e.parentStack[len(e.parentStack)-1]

	// Check if we have the same parent as the previous node.
	// EndElements are children of the StartElement/parent.
	// This means we are still on the same depth and can safely return here.
	if newNode.Parent == previousNode.Parent {
		// TODO: Handle when parent is nil or e.nodePath empty (for the first node only)
		return
	}

	// Is the previous node is our parent, we don't have to rebalance here
	if newNode.Parent == previousNode {
		return
	}

	// Does or stack match? This happens when we are the next node after an EndNode
	if newNode.Parent == lastStack {
		return
	}

	// Okk, now we have to work. The stack doesn't match the stack the newNode
	// expected.
	// TODO
	// 1. Get trees to the common parent
	leftTree, _, rightTree := getCommonPaths(newNode, e.parentStack)

	// 2. Add all EndNodes from the current stack till the root (rightTree)
	tmpParent := lastStack
	for i := range rightTree {
		elem := rightTree[i].Token.(xml.StartElement).End()
		e.nodePath = append(e.nodePath, &xmltree.Node{
			Token:  elem,
			Parent: tmpParent,
		})
		e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("EndNode(%s) - balanced", elem.Name.Local))
		tmpParent = rightTree[i].Parent
	}

	// 3. Add all StartNodes for the newNode stack till the root (leftTree)
	e.nodePath = append(e.nodePath, leftTree...)
	for i := range leftTree {
		e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("StartNode(%s) - balanced", leftTree[i].Token.(xml.StartElement).Name.Local))
	}

	// 4. TODO:
	// - How to (and if to) duplicate nodes that occure multiple times to have independent objects
	// - Verify parents are consistent
	//fmt.Printf("node: %T %v\n", newNode.Token, newNode.Token)
	//fmt.Printf("Propable missbalance at: \n\tstack:%v\n\t%v\n", e.parentStack, e.nodePathStr)
}

// getCommonPath returns the first node (from botton) that has the given node
// as one parent which is also present in the stack.
// leftTree holds the nodes that are parents of `node` up to the common root in reverse order.
// rightTree holds the nodes that are parents of `stack` up to the common root in reverse order.
func getCommonPaths(node *xmltree.Node, stack []*xmltree.Node) (leftTree []*xmltree.Node, commonParent *xmltree.Node, rightTree []*xmltree.Node) {
	// set empty slices instead of nils
	leftTree = []*xmltree.Node{}
	rightTree = []*xmltree.Node{}

	// get commonParent and build leftTree
nodeLoop:
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		for i := range stack {
			if parent == stack[i] {
				commonParent = parent
				break nodeLoop
			}
		}

		leftTree = append(leftTree, parent)
	}

	// build rightTree
	for i := range stack {
		missingNode := stack[len(stack)-1-i]
		if missingNode == commonParent {
			break
		}
		rightTree = append(rightTree, missingNode)
	}

	// reverse left and right tree, so the resulting trees are appendable to the stack
	reverseNodes(leftTree)
	reverseNodes(rightTree)

	return leftTree, commonParent, rightTree
}

// reverseNodes reverses xmltree.Node slices
func reverseNodes(nodes []*xmltree.Node) {
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
}
