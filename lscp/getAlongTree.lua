
local function getTypeOfKey(key) 
	return {key, redis.call("TYPE", key)["ok"]}
end

local queue = {getTypeOfKey(KEYS[1])}
local result = {}

while table.getn(queue) > 0 
do	
	if queue[1][2] == "hash" then
		local ret = redis.call("HGETALL", queue[1][1])
		local n = table.getn(ret)
		queue[1][3] = ret
		for i = 2, n, 2
		do
			local type = redis.call("TYPE", ret[i])["ok"]
			queue[1][3][i] = {ret[i], type}
			table.insert(queue, {ret[i], type})
		end
		table.insert(result, queue[1])
	
	elseif queue[1][2] == "list" then
		local ret = redis.call("LRANGE", queue[1][1], 0, -1)
		local n = table.getn(ret)
		queue[1][3] = {}
		for i=1, n, 1
		do
			local type = redis.call("TYPE", ret[i])["ok"]
			queue[1][3][i] = {ret[i], type}
			table.insert(queue, {ret[i], type})
		end
		table.insert(result, queue[1])

	elseif queue[1][2] == "string" then
		queue[1][3] = redis.call("GET", queue[1][1])

	elseif queue[1][2] == "set" then 
		local ret = redis.call("SMEMBERS", queue[1][1])
		local n = table.getn(ret)
		queue[1][3] = {}
		for i=1, n, 1
		do
			queue[1][3][i] = {ret[i]}
			table.insert(queue, queue[1][3][ret[i]])
		end
	end

	table.remove(queue, 1)
end

return result