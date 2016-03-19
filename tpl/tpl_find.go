package tpl

import (
	"github.com/LindenY/gomdise/lscp"
	"github.com/LindenY/gomdise/trans"
	"github.com/garyburd/redigo/redis"
	"fmt"
)

var (
	FindTemplate ActionTemplate = &findTemplate{}
)

type snode struct {
	parent   trans.RMNode
	children []trans.RMNode
	value    []interface{}
}

func (node *snode) Parent() trans.RMNode {
	return node.parent
}

func (node *snode) SetParent(parent trans.RMNode) {
	node.parent = parent
}

func (node *snode) Child(index int) trans.RMNode {
	return node.children[index]
}

func (node *snode) AddChildren(children ...trans.RMNode) {
	if children == nil {
		return
	}
	for _, child := range children {
		node.children = append(node.children, child)
		child.SetParent(node)
	}
}

func (node *snode) Size() int {
	return len(node.children)
}

func (node *snode) Value() interface{} {
	return node.value
}

/*
 *
 */
type findTemplate struct {
}

func (ft *findTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	action := &trans.Action{
		Name:    "GetAloneTree",
		Type:    trans.ScriptAction,
		Script:  lscp.LSGetAlongTree,
		Args:    redis.Args{args[0]},
		Handler: ft.handle,
	}
	*actions = append(*actions, action)
}

func (ft *findTemplate) handle(tran *trans.Transaction, action *trans.Action, reply interface{}) {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		panic(err)
	}
	root := &snode{
		children: make([]trans.RMNode, 0),
		value:    values(replies[0]),
	}
	currIdx := 1

	stack := make([]*snode, 0, len(replies) / 2)
	stack = append(stack, root)
	for len(stack) > 0 && currIdx < len(replies) {

		var init int
		var step int
		parent := popLast(&stack)

		switch parent.value[1].(string) {
		case "hash":
			init = 1
			step = 2
		case "list", "set", "zset":
			init = 0
			step = 1
		}

		for i := init; i < len(parent.value[2].([]interface{})); i = i + step {
			next := values(replies[currIdx])
			expected := parent.value[2].([]interface{})[i].([]interface{})[0]

			if next[0] != expected {
				continue
			}

			child := &snode{
				children: make([]trans.RMNode, 0),
				value:    next,
			}
			parent.AddChildren(child)
			stack = append(stack, child)
			currIdx++
		}
	}

	tran.State = root
	printRMNode(root, 0)
}

func values(rpy interface{}) []interface{} {
	elems, err := redis.Values(rpy, nil)
	if err != nil {
		panic(err)
	}

	retVal := make([]interface{}, len(elems), len(elems))
	retVal[0], err = redis.String(elems[0], err)
	retVal[1], err = redis.String(elems[1], err)
	if err != nil {
		panic(err)
	}

	if len(elems) < 3 {
		return retVal
	}

	init := 0
	step := 1
	max := len(elems[2].([]interface{}))
	val := make([]interface{}, max)

	if retVal[1] == "hash" {
		var err error
		for i := 0; i < max; i = i+2 {
			val[i], err = redis.String(elems[2].([]interface{})[i], err)
		}
		if err != nil {
			panic(err)
		}
		init = 1
		step = 2
	}

	for i := init; i < max; i=i+step {
		val[i] = values(elems[2].([]interface{})[i])
	}
	retVal[2] = val
	return retVal
}

func popLast(stack *[]*snode) *snode {
	l := len(*stack)
	last := (*stack)[l-1]
	*stack = (*stack)[0 : l-1]
	return last
}

func printRMNode(node trans.RMNode, indent int) {

	is := ""
	for i := 0; i < indent; i ++ {
		is += "\t"
	}
	fmt.Printf("%s %v \n", is, node)

	for i := 0; i < node.Size(); i++ {
		printRMNode(node.Child(i), indent + 1)
	}
}
