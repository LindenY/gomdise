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

	parent   *Action
	children []*Action
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
	child.parent = action
	action.children = append(action.children, child)
}

func (action *Action) Parent() RMNode {
	return action.parent
}

func (action *Action) Child(index int) RMNode {
	return action.children[index]
}

func (action *Action) Children(start int, end int) []RMNode {
	actions := action.children[start:end]
	nodes := make([]RMNode, len(actions))
	for i, a := range actions {
		nodes[i] = a
	}
	return nodes
}

func (action *Action) Size() int {
	return len(action.children)
}

func (action *Action) String() string {
	return fmt.Sprintf("%s\t%v", action.Name, action.Args)
}
