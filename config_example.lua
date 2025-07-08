local os = require("os")
local http = require("http")
local json = require("json")
-- Domains to monitor
dns.cloudflare[os.getenv("CF_TOKEN")] = {
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
dns.interval = 4 * time.duration.hour
