# dynamic-dns

> "What I cannot create, I do not understand" – Richard Feynman

I do not have affordable access to a static IPv4 address in my area. Like many
others I need a way to keep my DNS records updated to my IP. However, most of
these programs do not follow my philosophy of building many small useful tools
as a way of understanding more of the problem space.

`dynamic-dns` is my way of learning and building a dynamic DNS system. It uses
lua for its configuration to allow for user defined behavior and for a more
pleasurable configuration experience (i use neovim btw).

## Sample configuration
```lua
-- config.lua
local http = require("http")
local json = require("json")
-- Domains to monitor / update
dns.cloudflare["CF_TOKEN_CONTENTS"] = {
	"example.com",
}

function check_aws()
	local resp, status = http.get("https://checkip.amazonaws.com/", {})
	return resp, status
end

-- Define how to check IPv4 Address
dns.get_ipv4 = function()
	local ip, status = check_aws()
	if status == 200 then
		return ip
	end
	return ""
end

-- Interval to update the DNS records
dns.interval = 2 * time.duration.hour
```
