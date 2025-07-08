package lualibraries

import (
	"io"
	"net/http"

	"github.com/Shopify/go-lua"
	"github.com/Shopify/goluago/util"
)

var httpFunctions = []lua.RegistryFunction{
	{Name: "get", Function: func(l *lua.State) int {
		url := lua.CheckString(l, 1)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lua.Errorf(l, "unable to build new request: %s", err.Error())
			return 0
		}

		if !l.IsNil(2) {
			headers, err := util.PullStringTable(l, 2)
			if err != nil {
				lua.Errorf(l, "unable to acces headers table: %s", err.Error())
				return 0
			}
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lua.Errorf(l, "error fetching: %s", err.Error())
			return 0
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			lua.Errorf(l, "error reading body: %s", err.Error())
			return 0
		}
		l.PushString(string(b))
		l.PushInteger(resp.StatusCode)

		return 2
	}},
}

func LoadHTTP(l *lua.State, global bool) {
	lua.Require(l, "http", func(state *lua.State) int {
		lua.NewLibrary(state, httpFunctions)
		return 1
	}, global)
}
