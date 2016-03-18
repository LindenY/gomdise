package trans

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/LindenY/gomdise/lscp"
)

type ActionType int

const (
	CmdAction ActionType = iota
	ScriptAction
)

type Action struct {
	Name     string
	Script   *lscp.Script
	Type	 ActionType
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

func (action *Action) AddChildren(children ...*Action) {
	if children == nil {
		return
	}
	for _, child := range children {
		action.children = append(action.children, child)
		child.parent = action
	}
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

func (action *Action) Value() interface{} {
	return action.Reply
}

func (action *Action) String() string {

	return fmt.Sprintf("[%s:%v]\t%v", action.Name, action.Type, action.Args)
}
