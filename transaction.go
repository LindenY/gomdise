package gomdies

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type Transaction struct {
	conn    redis.Conn

	Actions []*Action
	Replies []interface{}
	Err     error
}

type Action struct {
	Name    string
	Args    redis.Args
	Handler ReplyHandler
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

func NewTransaction(pool *redis.Pool) *Transaction {
	t := &Transaction{
		conn: pool.Get(),
		Actions:make([]*Action, 0),
		Replies:make([]interface{}, 0),
	}
	return t
}

func (tran *Transaction) pushAction(action *Action) {
	tran.Actions = append(tran.Actions, action)
}

func (tran *Transaction) popAction() *Action {
	if (len(tran.Actions) == 0) {
		return nil
	}
	ret := tran.Actions[len(tran.Actions) - 1]
	tran.Actions = tran.Actions[0:len(tran.Actions) - 1]
	return ret
}

func (tran *Transaction) setError(err error) {
	if tran.Err == nil {
		tran.Err = err
	}
}

func (tran *Transaction) sendAction(action *Action) error {
	fmt.Printf("send action: %v \n", action)
	return tran.conn.Send(action.Name, action.Args...)
}

func (tran *Transaction) doAction(action *Action) (interface{}, error) {
	fmt.Printf("do action: %v \n", action)
	return tran.conn.Do(action.Name, action.Args...)
}

func (tran *Transaction) exec(node *MNode) {
	if len(tran.Actions) == 1 {
		reply, err := tran.doAction(tran.Actions[0])
		if err != nil {
			panic(err)
		}
		tran.handle(node, reply)
		return
	}

	if err := tran.conn.Send("MULTI"); err != nil {
		panic(err)
	}
	for _, action := range tran.Actions {
		if err := tran.sendAction(action); err != nil {
			panic(err)
		}
	}
	reply, err := tran.conn.Do("EXEC")
	if err != nil {
		panic(err)
	}
	tran.handle(node, reply)
}

func (tran *Transaction) handle(node *MNode, reply interface{}) {
	var replies []interface{}
	if len(tran.Actions) == 1 {
		replies = []interface{}{reply}
	} else {
		var err error
		replies, err = redis.Values(reply, nil)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Num Action:%d, Num Replies:%d \n", len(tran.Actions), len(replies))
	for i, action := range tran.Actions {
		fmt.Printf("\t[%d]: %v \n", i, action)
		fmt.Printf("\t \t %v \n", replies[i])
		action.handle(tran, replies[i])
	}
}

func (tran *Transaction) Exec() *MNode {
	defer tran.conn.Close()


	root := newMNode(nil)
	for len(tran.Actions) > 0 {
		tran.exec(root)

		next := make([]*Action, 0)
		for _, action := range tran.Actions {
			if len(action.Children) > 0 {
				next = append(next, action.Children...)
			}
		}
		tran.Actions = next
		fmt.Println(len(tran.Actions))
	}
	return nil
}
