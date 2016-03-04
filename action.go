package gomdies

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type Action struct {
	Name     string
	Args     redis.Args
	Reply    interface{}
	Handler  ReplyHandler
	Children []*Action
}

type ReplyHandler func(tran *Transaction, action *Action, reply interface{})

func (action *Action) handle(trans *Transaction, reply interface{}) {
	if action.Handler != nil {
		action.Handler(trans, action, reply)
	}
}

func (action *Action) addChild(child *Action) {
	if child == nil {
		return
	}
	action.Children = append(action.Children, child)
}

func (action *Action) String() string {
	return fmt.Sprintf("%s\t%v", action.Name, action.Args)
}
