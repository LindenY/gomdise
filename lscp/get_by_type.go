package lscp

var LSGetByType Script = NewScript(
	1,
	`local key = KEYS[1]
	local type = redis.call("TYPE", key)["ok"]
	local ret
	if type == "string" then
		ret = redis.call("GET", key)
	end
	if type == "hash" then
		ret = redis.call("HGETALL", key)
	end
	if type == "list" then
		ret = redis.call("LRANGE", key)
	end
	if type == "set" then
		ret = redis.call("SMEMBERS", key)
	end
	return {type, ret}`)
