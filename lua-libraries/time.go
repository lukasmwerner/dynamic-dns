package lualibraries

import "github.com/Shopify/go-lua"

func LoadTime(l *lua.State, global bool) {
	lua.Require(l, "time", func(l *lua.State) int {
		l.NewTable() // duration
		l.NewTable() // durations

		nanosecond := 1
		l.PushInteger(nanosecond)
		l.SetField(-2, "nanosecond")

		microsecond := 1000 * nanosecond
		l.PushInteger(microsecond)
		l.SetField(-2, "microsecond")

		millisecond := 1000 * microsecond
		l.PushInteger(millisecond)
		l.SetField(-2, "millisecond")

		second := 1000 * millisecond
		l.PushInteger(second)
		l.SetField(-2, "second")

		minute := 60 * second
		l.PushInteger(minute)
		l.SetField(-2, "minute")

		hour := 60 * minute
		l.PushInteger(hour)
		l.SetField(-2, "hour")

		day := 24 * hour
		l.PushInteger(day)
		l.SetField(-2, "day")

		l.SetField(-2, "duration")
		return 1
	}, global)
}
