package cpxtemplate

import (
	"encoding/xml"
	"fmt"
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
func (e *LuaEngine) GetNodePathString() []string {
	return e.nodePathStr
}

func (e *LuaEngine) iStartNode(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)

	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("StartNode(%d)", nodeId))

	// Append node to the nodePath
	node := e.lt.NodeList[nodeId]
	e.nodePath = append(e.nodePath, node)

	// Add node to the parent stack
	e.parentStack = append(e.parentStack, node)

	// TODO
	// - If this nodes parent is different than the one on the stack, add EndNode of the old parent
	return 0
}

func (e *LuaEngine) iEndNode(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)

	// Append node to the nodePath
	node := e.lt.NodeList[nodeId]
	e.nodePath = append(e.nodePath, node)

	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("EndNode(%d)", nodeId))

	// TODO
	// - Remove one level from the parent stack
	return 0
}

func (e *LuaEngine) iSetToken(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)

	// Append node to the nodePath
	node := e.lt.NodeList[nodeId]
	e.nodePath = append(e.nodePath, node)

	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("SetToken(%d)", nodeId))

	// TODO
	// - If parent (=original start call) was not on the stack: Do what?
	return 0
}

func (e *LuaEngine) iCharData(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)

	// Append node to the nodePath
	node := e.lt.NodeList[nodeId]
	e.nodePath = append(e.nodePath, node)

	e.nodePathStr = append(e.nodePathStr, fmt.Sprintf("CharData(%d)", nodeId))

	// TODO
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
	sc.WriteString("\n")

	e.nodePathStr = append(e.nodePathStr, "Print(???)")

	// Create new node and add it to the nodePath
	chrNode := xml.CharData(sc.String())
	node := &xmltree.Node{
		Token:  chrNode,
		Parent: e.parentStack[len(e.parentStack)-1],
	}
	e.nodePath = append(e.nodePath, node)

	// TODO
	return 0
}

/*
<p1>
  <p2>[[ for A={1..2} ]]</p2>
  <p3>[# A #]</p3>
  <p4>[[ end ]]</p4>
</p1>

## Loop case

StartNode(p1)
 StartNode(p2)
  for A={1..2}
  StartNode(p3)
   Print(A)
   StartNode(p4)
    endfor
    EndNode(p4)
   EndNode(p3)
  EndNode(p2)
 EndNode(p1)

StartNode(p1)
 StartNode(p2)
  StartNode(p3)
   Print(A)
   StartNode(p4)
  StartNode(p3) -- visited 2nd time. So we end the last p3 before it?
   Print(A)
   StartNode(p4)
    EndNode(p4)
   EndNode(p3)
  EndNode(p2)
 EndNode(p1)

## Non loop case
StartNode(p1)
 StartNode(p2)
  if False
  StartNode(p3)
   Print(A)
   StartNode(p4)
    endif
    EndNode(p4)
   EndNode(p3)
  EndNode(p2)
 EndNode(p1)

StartNode(p1)
 StartNode(p2)
    EndNode(p4) -- We end a non started node. So we start all parents until we have a balance?
   EndNode(p3)
  EndNode(p2)
 EndNode(p1)

## Loop higher case
StartNode(p1)
 StartNode(p2) -- loop target
  for A={1..2}
  StartNode(p3)
   Print(A)
   StartNode(p4)
    endfor
    EndNode(p4)
   EndNode(p3)
  EndNode(p2)
 EndNode(p1)

StartNode(p1)
 StartNode(p2) -- Loop target
  StartNode(p3)
   Print(A)
   StartNode(p4)
  StartNode(p3) -- visited 2nd time. So we insert each parent as StartNode til the loop target
   Print(A)
   StartNode(p4)
    EndNode(p4)
   EndNode(p3)
  EndNode(p2)
 EndNode(p1)
*/
