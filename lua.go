package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Shopify/go-lua"
	lualibraries "github.com/lukasmwerner/dynamic-dns/lua-libraries"
)

func LoadLua(fileName string) *lua.State {
	l := lua.NewStateEx()
	lualibraries.Open(l)
	err := lua.DoFile(l, fileName)
	if err != nil {
		panic(err.Error())
	}

	return l
}

func FetchConfigIPv4(l *lua.State) (string, error) {
	l.Global("dns")
	l.Field(-1, "get_ipv4")
	l.Call(0, 1)
	ip, ok := l.ToString(-1)
	l.Pop(2)
	if !ok {
		return "", fmt.Errorf("unable to get IPv4 address from lua via dns.get_ipv4")
	}
	if ip == "" {
		return "", fmt.Errorf("lua unable to get IPv4 address in dns.get_ipv4")
	}
	ip = strings.TrimSpace(ip)
	return ip, nil
}

type Notification struct {
	Domain string
	OldIP  string
	NewIP  string
}

func Notify(l *lua.State, n Notification) {
	l.Global("dns")
	if !l.IsTable(-1) {
		l.Pop(1)
		log.Println("error: global 'dns' is not a table")
		return
	}

	l.Field(-1, "notify")
	if !l.IsFunction(-1) && !l.IsNil(-1) {
		l.Pop(2)
		log.Println("error: 'dns.notify' is not a function")
		return
	} else if l.IsNil(-1) {
		l.Pop(2)
		return
	}

	l.Remove(-2)

	l.PushString(n.Domain)
	l.PushString(n.OldIP)
	l.PushString(n.NewIP)

	l.Call(3, 0)
}

func GetConfigInterval(l *lua.State) time.Duration {
	l.Global("dns")
	l.Field(-1, "interval")
	interval := lua.CheckInteger(l, -1)
	l.Pop(1)
	return time.Duration(interval)
}

func GetCloudflareConfigData(l *lua.State) map[string][]string {
	m := make(map[string][]string)

	l.Global("dns")
	l.Field(1, "cloudflare")
	l.PushNil()      // Need an extra slot for next to push the value
	for l.Next(-2) { // Stack: [dns, cloudflare, key, nil/list]
		// key is at -2
		// value is at -1
		key := lua.CheckString(l, -2)

		if !l.IsTable(-1) {
			continue
		}

		values := []string{}
		l.PushNil()
		for l.Next(-2) {
			if !l.IsString(-1) {
				continue
			}
			values = append(values, lua.CheckString(l, -1))
			l.Pop(1)
		}

		m[key] = values

		l.Pop(1) // Make sure that next is on the value of the table
	}
	l.Pop(2)
	return m
}

func GetPorkbunConfigData(l *lua.State) map[KeyPair][]string {
	m := make(map[KeyPair][]string)

	l.Global("dns")
	l.Field(1, "porkbun")
	l.PushNil()      // Need an extra slot for next to push the value
	for l.Next(-2) { // Stack: [dns, porkbun, key, nil/list]
		// key is at -2
		// value is at -1
		if !l.IsTable(-2) {
			continue
		}

		if !l.IsTable(-1) {
			continue
		}

		values := []string{}
		l.PushNil()
		for l.Next(-2) {
			if !l.IsString(-1) {
				continue
			}
			values = append(values, lua.CheckString(l, -1))
			l.Pop(1)
		}
		l.Pop(1)

		key_parts := []string{}
		l.PushNil()
		for l.Next(-2) {
			if !l.IsString(-1) {
				continue
			}
			key_parts = append(key_parts, lua.CheckString(l, -1))
			l.Pop(1)
		}
		l.PushNil() // Put the 'Value' back on the stack

		m[KeyPair{
			Key:    key_parts[0],
			Secret: key_parts[1],
		}] = values

		l.Pop(1) // Make sure that next is on the value of the table
	}
	l.Pop(2)
	return m
}

func IsHttpAPIEnabled(l *lua.State) bool {
	l.Global("dns")
	l.Field(-1, "api")
	enabled := l.ToBoolean(-1)
	l.Pop(1)
	return enabled
}
