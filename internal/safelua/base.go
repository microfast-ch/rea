// Package safelua adds some safe functions from the Shopify go-lua base library.
// https://github.com/Shopify/go-lua/blob/main/base.go
package safelua

import (
	"strconv"
	"strings"

	"github.com/Shopify/go-lua"
)

func Add(l *lua.State) {
	l.Register("next", next)
	l.Register("pairs", pairs("__pairs", false, next))
	l.Register("ipairs", pairs("__ipairs", true, intPairs))
	l.Register("tonumber", tonumber)

	l.Register("getmetatable", func(l *lua.State) int {
		lua.CheckAny(l, 1)
		if !l.MetaTable(1) {
			l.PushNil()
			return 1
		}
		lua.MetaField(l, 1, "__metatable")
		return 1
	})

	l.Register("setmetatable", func(l *lua.State) int {
		t := l.TypeOf(2)
		lua.CheckType(l, 1, lua.TypeTable)
		lua.ArgumentCheck(l, t == lua.TypeNil || t == lua.TypeTable, 2, "nil or table expected")
		if lua.MetaField(l, 1, "__metatable") {
			lua.Errorf(l, "cannot change a protected metatable") // TODO: Test if Errorf works for us and how
		}
		l.SetTop(2)
		l.SetMetaTable(1)
		return 1
	})

	l.Register("tostring", func(l *lua.State) int {
		lua.CheckAny(l, 1)
		lua.ToStringMeta(l, 1)
		return 1
	})

	l.Register("type", func(l *lua.State) int {
		lua.CheckAny(l, 1)
		l.PushString(lua.TypeNameOf(l, 1))
		return 1
	})
}

func next(l *lua.State) int {
	lua.CheckType(l, 1, lua.TypeTable)
	l.SetTop(2)
	if l.Next(1) {
		return 2
	}
	l.PushNil()
	return 1
}

func pairs(method string, isZero bool, iter lua.Function) lua.Function {
	return func(l *lua.State) int {
		if hasMetamethod := lua.MetaField(l, 1, method); !hasMetamethod {
			lua.CheckType(l, 1, lua.TypeTable) // argument must be a table
			l.PushGoFunction(iter)             // will return generator,
			l.PushValue(1)                     // state,
			if isZero {                        // and initial value
				l.PushInteger(0)
			} else {
				l.PushNil()
			}
		} else {
			l.PushValue(1) // argument 'self' to metamethod
			l.Call(1, 3)   // get 3 values from metamethod
		}
		return 3
	}
}

func tonumber(l *lua.State) int {
	if l.IsNoneOrNil(2) { // standard conversion
		if n, ok := l.ToNumber(1); ok {
			l.PushNumber(n)
			return 1
		}
		lua.CheckAny(l, 1)
	} else {
		s := lua.CheckString(l, 1)
		base := lua.CheckInteger(l, 2)
		lua.ArgumentCheck(l, 2 <= base && base <= 36, 2, "base out of range")
		if i, err := strconv.ParseInt(strings.TrimSpace(s), base, 64); err == nil {
			l.PushNumber(float64(i))
			return 1
		}
	}
	l.PushNil()
	return 1
}

func intPairs(l *lua.State) int {
	i := lua.CheckInteger(l, 2)
	lua.CheckType(l, 1, lua.TypeTable)
	i++ // next value
	l.PushInteger(i)
	l.RawGetInt(1, i)
	if l.IsNil(-1) {
		return 1
	}
	return 2
}
