package lscp

import "github.com/garyburd/redigo/redis"

type Script struct {
	rscp *redis.Script
	loaded bool
}

func NewScript(keyCount int, src string) *Script {
	rscp := redis.NewScript(keyCount, src)
	lscp := new(Script)
	lscp.loaded = false
	lscp.rscp = rscp
	return lscp
}

func (scp *Script) Send(c redis.Conn, args ...interface{}) error {
	if scp.loaded {
		return scp.rscp.SendHash(c, args...)
	} else {
		scp.loaded = true
		return scp.rscp.Send(c, args...)
	}
}

func (scp *Script) Do(c redis.Conn, args ...interface{}) (interface{}, error) {
	if !scp.loaded {
		scp.rscp.Load(c)
	}
	return scp.rscp.Do(c, args...)
}