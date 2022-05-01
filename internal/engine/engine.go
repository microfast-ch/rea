package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/Shopify/go-lua"
	"github.com/djboris9/xmltree"
	"github.com/microfast-ch/rea/internal/safelua"
	"golang.org/x/exp/slices"

	goluagoUtil "github.com/Shopify/goluago/util"
)

type LuaEngine struct {
	lt       *LuaTree
	luaState *lua.State

	// State information during exec
	execLock    sync.Mutex
	nodePath    []*xmltree.Node
	parentStack []*xmltree.Node // stack of StartNode elements to keep the balance
	nodePathStr []string

	// Counts how many times the function at the line of code was executed.
	// This is needed to keep track for the iteration target.
	reachcounter *reachcounter

	// List of xml node names that act as the origin of iterations
	iterationNodes []string
}

// Passed data must be a primitive or a map.
type TemplateData struct {
	Data     map[string]any
	Metadata map[string]string
}

func NewLuaEngine(lt *LuaTree, data *TemplateData) *LuaEngine {
	// Initialize lua
	l := lua.NewState()
	// lua.BaseOpen(l) // This should be uncommented for debugging purposes only

	e := &LuaEngine{
		lt:           lt,
		luaState:     l,
		reachcounter: newReachCounter(),
	}

	// Map our Go functions to lua
	l.Register("SetToken", e.handleIterations(e.iSetToken))
	l.Register("StartNode", e.handleIterations(e.iStartNode))
	l.Register("EndNode", e.handleIterations(e.iEndNode))
	l.Register("CharData", e.handleIterations(e.iCharData))
	l.Register("Print", e.handleIterations(e.iPrint))
	l.Register("SetIterationNodes", e.iSetIterationNodes)

	// Inject data into the lua stack
	if data != nil {
		if data.Data != nil {
			for k, v := range data.Data {
				goluagoUtil.DeepPush(l, v)
				l.SetGlobal(k)
			}
		}

		if data.Metadata != nil {
			goluagoUtil.DeepPush(l, data)
			l.SetGlobal("metadata")
		}
	}

	// Restricted base library
	safelua.Add(l)

	// Return engine
	return e
}

// This function is serialized on exec.
func (e *LuaEngine) Exec(initFunc string) error {
	e.execLock.Lock() // TODO: We might convert it to sync.Once
	defer e.execLock.Unlock()

	// Initialize empty state
	e.nodePath = []*xmltree.Node{}
	e.nodePathStr = []string{}

	// Execute initialization function
	err := lua.DoString(e.luaState, initFunc)
	if err != nil {
		// We got an error, but the detailed error message is on the stack. Wrap it.
		return fmt.Errorf("executing init lua prog got %w with :%s", err, lua.CheckString(e.luaState, -1))
	}

	// Execute lua program
	err = lua.DoString(e.luaState, e.lt.LuaProg)
	if err != nil {
		// We got an error, but the detailed error message is on the stack. Wrap it.
		return fmt.Errorf("executing lua prog got %w with :%s", err, lua.CheckString(e.luaState, -1))
	}

	return err
}

// Can only be called after Exec() has been run.
func (e *LuaEngine) WriteXML(w io.Writer) error {
	enc := xml.NewEncoder(w)
	for i := range e.nodePath {
		// if err := EncodeToken(enc, w, e.nodePath[i].Token); err != nil { // Is a custom encoder needed
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

// Can only be called after Exec() has been run.
func (e *LuaEngine) GetNodePath() []*xmltree.Node {
	return e.nodePath
}

// Can only be called after Exec() has been run.
func (e *LuaEngine) GetNodePathString() []string {
	return e.nodePathStr
}

func (e *LuaEngine) iStartNode(state *lua.State) int {
	nodeID := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeID]

	// Fill tree, before we add us to the parent stack
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("StartNode(%d)", nodeID))

	// Add node to the parent stack
	e.parentStack = append(e.parentStack, node)

	return 0
}

func (e *LuaEngine) iEndNode(state *lua.State) int {
	nodeID := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeID]

	// Fill tree, before we remove the last parent from the stack
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("EndNode(%d)", nodeID))

	// Remove one level from the parent stack
	e.parentStack = e.parentStack[:len(e.parentStack)-1]

	return 0
}

func (e *LuaEngine) iSetToken(state *lua.State) int {
	nodeID := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeID]

	// Fill tree, before we are added to the path
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("SetToken(%d)", nodeID))

	return 0
}

func (e *LuaEngine) iCharData(state *lua.State) int {
	nodeID := lua.CheckInteger(state, -1)
	node := e.lt.NodeList[nodeID]

	// Fill tree, before we are added to the path
	e.fillTree(node)

	// Append node to the nodePath
	e.nodePath = append(e.nodePath, node)
	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("CharData(%d)", nodeID))

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

func (e *LuaEngine) iSetIterationNodes(state *lua.State) int {
	// Extract string array from first argument, which is a table
	idx := state.AbsIndex(-1)
	args := make([]string, lua.LengthEx(state, idx))

	state.PushNil()

	for state.Next(idx) {
		k, ok := state.ToInteger(-2)
		if !ok {
			lua.Errorf(state, "SetIterationNodes cannot process numeric index, got: %s", state.TypeOf(-2))
			panic("unreachable")
		}

		args[k-1] = lua.CheckString(state, -1)
		state.Pop(1)
	}

	// Set new iteration nodes
	e.SetIterationNodes(args)

	return 0
}

// SetIterationNodes updates the list of node names that act as an iteration origin.
func (e *LuaEngine) SetIterationNodes(nodes []string) {
	e.iterationNodes = nodes
}

// handleIterations is a middleware for handling iterations. It detects if we
// need to reconstruct XML parents from the iteration origin up to the current node.
// This function needs to be called before processing elements that have an impact
// on the resulting nodePath structure.
func (e *LuaEngine) handleIterations(next lua.Function) lua.Function {
	return func(state *lua.State) int {
		wasPrevious := e.countCall(state)

		// If we are processing this node for the first time, just continue
		if !wasPrevious {
			return next(state)
		}

		// We already saw this node. So check if we have an iteration origin in the parent tree
		lastNode := e.nodePath[len(e.nodePath)-1]
		var iterOrigin *xmltree.Node

		for parent := lastNode.Parent; parent.Token != nil; parent = parent.Parent {
			elem := parent.Token.(xml.StartElement)
			if slices.Contains(e.iterationNodes, elem.Name.Local) {
				iterOrigin = parent
			}
		}

		// If we haven't found an iterOrigin, we are not in an iteration context
		if iterOrigin == nil {
			return next(state)
		}

		// Rebalance tree up to the iterTarget. The next function will rebalance
		// it again down to the new node.
		e.fillTree(iterOrigin)

		// Clean counter and add current node to it, as it will be rendered
		e.reachcounter.Clean()
		e.countCall(state)

		return next(state)
	}
}

// countCall records the execution of the current line and returns true if this
// function was already called at least one time at the current source line.
// TODO: If the function is defined two times at the same line, it currently
// cannot distinguish it.
func (e *LuaEngine) countCall(state *lua.State) bool {
	lua.Where(state, 1)

	s, ok := state.ToString(-1)
	if !ok {
		panic("invalid lua state reached in countCall")
	}

	state.Pop(1)

	i := e.reachcounter.Add(s)

	return i != 0
}

// On each node this function is called. It checks if the new node has the same
// parent as the previous one. If not, it determines the common parent and closes
// the tags parent tags up to it on the old branch. Then it opens all start tags of the new
// branch so the new tag can be added.
// This needs to be called before the node is added to the path.
// nolint:funlen
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
		return
	}

	// Is the previous node is our parent, we don't have to rebalance here
	if newNode.Parent == previousNode {
		return
	}

	// Does our stack match? This happens when we are the next node after an EndNode
	if newNode.Parent == lastStack {
		return
	}

	// Okk, now we have to work. The stack doesn't match the stack the newNode
	// expected.

	// Get trees to the common parent
	leftTree, _, rightTree := getCommonPaths(newNode, e.parentStack)

	// Now handle the special case where we have excessive tags in simple logic blocks.
	// This is a heuristic to render the following block to `<body><p></p></body>` instead of `<body><p></p><p></p></body>`.
	//   <body><p>[[ if false then ]]</p><p>Hello</p><p>[[ end ]]</p></body>
	// It prevents inserting of balancing nodes if the newNode will establish the balance anyways.
	// Such simple blocks need to have only one level to balance and the nextToken must match the to-be-balanced nodes.
	if len(leftTree) == 1 && len(rightTree) == 1 {
		leftToken := leftTree[0].Token.(xml.StartElement).Name.Local
		rightToken := rightTree[0].Token.(xml.StartElement).Name.Local // This would be converted to an EndElement in balancing

		// The nextToken needs to be an EndElement to stay in balance
		nextToken, ok := newNode.Token.(xml.EndElement)
		if ok && leftToken == rightToken && leftToken == nextToken.Name.Local {
			// We got a match and have a group of same to-be-balanced nodes and the nextToken
			// fits into this scheme like `</a><a></a>`
			return
		}
	}

	// Add all EndNodes from the current stack till the root (rightTree) in reverse order.
	// This closes every node we started so we can create new nodes and stay in balance.
	tmpParent := lastStack

	for revIdx := range rightTree {
		i := len(rightTree) - revIdx - 1
		elem := rightTree[i].Token.(xml.StartElement).End()

		e.nodePath = append(e.nodePath, &xmltree.Node{
			Token:  elem,
			Parent: tmpParent,
		})
		e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("EndNode(%s) - balanced", elem.Name.Local))
		tmpParent = rightTree[i].Parent

		// Remove one level from the parent stack to keep the parent stack up to date
		e.parentStack = e.parentStack[:len(e.parentStack)-1]
	}

	// Add all StartNodes for the newNode stack till the root (leftTree).
	// This reconstructs our parent path so we will stay on the same level as before
	// and can proceed with the execution.
	e.nodePath = append(e.nodePath, leftTree...)
	for i := range leftTree {
		e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("StartNode(%s) - balanced", leftTree[i].Token.(xml.StartElement).Name.Local))

		// Add node to the parent stack to keep the parent stack up to date
		e.parentStack = append(e.parentStack, leftTree[i])
	}
}

// getCommonPath returns the first node (from botton) that has the given node
// as one parent which is also present in the stack.
// leftTree holds the nodes that are parents of `node` up to the common root in reverse order.
// rightTree holds the nodes that are parents of `stack` up to the common root in reverse order.
func getCommonPaths(node *xmltree.Node, stack []*xmltree.Node) (leftTree []*xmltree.Node,
	commonParent *xmltree.Node, rightTree []*xmltree.Node) {
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

// reverseNodes reverses xmltree.Node slices.
func reverseNodes(nodes []*xmltree.Node) {
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
}
