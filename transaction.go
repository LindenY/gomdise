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


type ReplyHandler func(tran *Transaction, action *Action, reply interface{}) error

func (action *Action) handle(trans *Transaction, reply interface{}) error {
	if action.Handler != nil {
		return action.Handler(trans, action, reply)
	}
	return nil
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

func (tran *Transaction) exec(node *mnode) error {
	if len(tran.Actions) == 1 {
		reply, err := redis.Values(tran.doAction(tran.Actions[0]))
		if err != nil {
			return err
		}
		fmt.Printf("reply: %v \n", reply)
		tran.Replies = append(tran.Replies, reply)
		return tran.Actions[0].handle(tran, reply)
	}

	if err := tran.conn.Send("MULTI"); err != nil {
		return err
	}
	for _, action := range tran.Actions {
		if err := tran.sendAction(action); err != nil {
			return err
		}
	}

	replies, err := redis.Values(tran.conn.Do("EXEC"))
	if err != nil {
		return err
	}

	fmt.Printf("reply: %v \n", replies)
	for i, reply := range replies {
		fmt.Printf("\t[%d]: %v \n", i, reply)

		if err = tran.Actions[i].handle(tran, reply); err != nil {
			break
		}
	}
	return err
}

func (tran *Transaction) Exec() *mnode {
	defer tran.conn.Close()


	root := newMNode(nil)
	for len(tran.Actions) > 0 {
		if err := tran.exec(root); err != nil {
			return err
		}

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
