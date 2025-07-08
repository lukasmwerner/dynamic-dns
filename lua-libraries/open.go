package lualibraries

import "github.com/Shopify/go-lua"

func Open(l *lua.State) {
	// standard lua libraries
	lua.Require(l, "_G", lua.BaseOpen, true)
	l.Pop(1)
	lua.Require(l, "package", lua.PackageOpen, true)
	l.Pop(1)
	lua.Require(l, "string", lua.StringOpen, true)
	l.Pop(1)
	lua.Require(l, "table", lua.TableOpen, true)
	l.Pop(1)
	lua.Require(l, "math", lua.MathOpen, true)
	l.Pop(1)

	// custom libraries
	LoadOS(l, false)
	l.Pop(1)
	LoadHTTP(l, false)
	l.Pop(1)
	LoadJSON(l, false)
	l.Pop(1)
	LoadTime(l, true)
	l.Pop(1)
	LoadDNS(l)
	l.Pop(1)
}
