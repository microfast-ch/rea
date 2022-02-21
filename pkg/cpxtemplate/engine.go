package cpxtemplate

import (
	"fmt"

	"github.com/Shopify/go-lua"
)

type LuaEngine struct {
	lt       *LuaTree
	luaState *lua.State
}

func NewLuaEngine(lt *LuaTree) *LuaEngine {
	// Initialize lua
	l := lua.NewState()
	lua.OpenLibraries(l)

	e := &LuaEngine{
		lt:       lt,
		luaState: l,
	}

	// Map our Go functions to lua
	l.Register("SetToken", e.iSetToken)
	l.Register("StartNode", e.iStartNode)
	l.Register("EndNode", e.iEndNode)
	l.Register("CharData", e.iCharData)

	// Return engine
	return e
}

func (e *LuaEngine) Exec() error {
	// Execute lua program
	err := lua.DoString(e.luaState, e.lt.LuaProg)
	if err != nil {
		// We got an error, but the detailed error message is on the stack. Wrap it.
		return fmt.Errorf("executing lua prog got %w with :%s", err, lua.CheckString(e.luaState, -1))
	}
	return err
}

func (e *LuaEngine) iStartNode(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)
	//node := e.lt.NodeList[nodeId]

	fmt.Printf("StartNode(%d) called\n", nodeId)

	// TODO
	return 0
}

func (e *LuaEngine) iEndNode(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)
	fmt.Printf("EndNode(%d) called\n", nodeId)
	// TODO
	return 0
}

func (e *LuaEngine) iSetToken(state *lua.State) int {
	nodeId := lua.CheckInteger(state, -1)
	fmt.Printf("SetToken(%d) called\n", nodeId)
	// TODO
	return 0
}

func (e *LuaEngine) iCharData(state *lua.State) int {
	fmt.Println("CharData")
	// TODO
	return 0
}
