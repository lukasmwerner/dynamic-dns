package lualibraries

import (
	"os"

	"github.com/Shopify/go-lua"
)

var osFunctions = []lua.RegistryFunction{
	{Name: "getenv", Function: func(l *lua.State) int {
		key := lua.CheckString(l, 1)
		value := os.Getenv(key)
		l.PushString(value)
		return 1
	}},
}

func LoadOS(l *lua.State, global bool) {
	lua.Require(l, "os", func(l *lua.State) int {
		lua.NewLibrary(l, osFunctions)
		return 1
	}, global)
}
