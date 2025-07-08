package lualibraries

import (
	"encoding/json"

	"github.com/Shopify/go-lua"
	"github.com/Shopify/goluago/util"
)

var jsonFunctions = []lua.RegistryFunction{
	{Name: "decode", Function: func(l *lua.State) int {
		payload := lua.CheckString(l, 1)
		var output any
		if err := json.Unmarshal([]byte(payload), &output); err != nil {
			lua.Errorf(l, "error parsing json: %s", err.Error())
			return 0
		}
		return util.DeepPush(l, output)
	}},
}

func LoadJSON(l *lua.State, global bool) {
	lua.Require(l, "json", func(l *lua.State) int {
		lua.NewLibrary(l, jsonFunctions)
		return 1
	}, global)
}
