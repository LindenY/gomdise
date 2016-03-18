
local function getByType(key) 
	local type = redis.call("TYPE", key)["ok"]
	return {key, type}
end

local t = { }
local key = KEYS[1]

t.key = key

return t