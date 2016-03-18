package lscp

var LSGetAlongTree *Script = NewScript(1, `
	local queue = { {KEYS[1], redis.call("TYPE", KEYS[1])["ok"]} }
	local visited = {}
	local result = {}

	while table.getn(queue) > 0
	do
		if not visited[queue[1][1]] then
			visited[queue[1][1]] = true

			if queue[1][2] == "hash" then
				local ret = redis.call("HGETALL", queue[1][1])
				local n = table.getn(ret)
				queue[1][3] = ret
				for i = 2, n, 2
				do
					local type = redis.call("TYPE", ret[i])["ok"]
					table.insert(queue, {ret[i], type})
					queue[1][3][i] = {ret[i], type}
				end
				table.insert(result, queue[1])

			elseif queue[1][2] == "string" then
				queue[1][3] = redis.call("GET", queue[1][1])
				table.insert(result, queue[1])

			elseif queue[1][2] == "set"
					or queue[1][2] == "list"
					or queue[1][2] == "zset" then
				local ret
				if queue[1][2] == "set" then
					ret = redis.call("SMEMBERS", queue[1][1])
				elseif queue[1][2] == "list" then
					ret = redis.call("LRANGE", queue[1][1], 0, -1)
				elseif queue[1][2] == "zset" then
					ret = redis.call("ZRANGE", queue[1][1], 0, -1)
				end

				local n = table.getn(ret)
				queue[1][3] = {}
				for i=1, n, 1
				do
					local type = redis.call("TYPE", ret[i])["ok"]
					table.insert(queue, {ret[i], type})
					queue[1][3][i] = {ret[i], type}
				end
				table.insert(result, queue[1])
			end
		end

		table.remove(queue, 1)
	end

	return result
`)

