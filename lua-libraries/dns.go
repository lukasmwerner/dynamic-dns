package lualibraries

import (
	"net"

	"github.com/Shopify/go-lua"
)

var dnsFunctions = []lua.RegistryFunction{{
	Name: "lookup", Function: func(l *lua.State) int {
		domain := lua.CheckString(l, -1)
		addrs, err := net.LookupHost(domain)
		if err != nil {
			lua.Errorf(l, "error looking up %s: %s", domain, err.Error())
		}
		if len(addrs) == 0 {
			lua.Errorf(l, "error looking up %s: %s", domain, "no results")
		}
		l.PushString(addrs[0])
		return 1
	},
}}

func LoadDNS(l *lua.State) {
	l.NewTable() // DNS table
	l.NewTable() // Cloudflare table
	l.SetField(-2, "cloudflare")
	lua.SetFunctions(l, dnsFunctions, 0)
	l.SetGlobal("dns")

	l.PushInteger(1) // HACK: make it so we can pop something...
}
