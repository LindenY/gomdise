package trans

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type Transaction struct {
	conn    redis.Conn
	Actions []*Action
	Err     error
}

func NewTransaction(pool *redis.Pool) *Transaction {
	t := &Transaction{
		conn:    pool.Get(),
		Actions: make([]*Action, 0),
	}
	return t
}

func (tran *Transaction) pushAction(action *Action) {
	tran.Actions = append(tran.Actions, action)
}

func (tran *Transaction) popAction() *Action {
	if len(tran.Actions) == 0 {
		return nil
	}
	ret := tran.Actions[len(tran.Actions)-1]
	tran.Actions = tran.Actions[0 : len(tran.Actions)-1]
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

func (tran *Transaction) exec() {
	if len(tran.Actions) == 1 {
		reply, err := tran.doAction(tran.Actions[0])
		if err != nil {
			panic(err)
		}
		tran.handle(reply)
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
	tran.handle(reply)
}

func (tran *Transaction) handle(reply interface{}) {
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
	prev := tran.Actions
	tran.Actions = make([]*Action, 0)
	for i, action := range prev {
		fmt.Printf("\t[%d]: %v \n", i, action)
		fmt.Printf("\t \t %v \n", replies[i])

		action.Reply = replies[i]
		action.handle(tran, replies[i])
	}
}

func (tran *Transaction) Exec() {
	defer tran.conn.Close()
	fmt.Printf("Trans starts executing")
	for len(tran.Actions) > 0 {
		tran.exec()

		/*
		next := make([]*Action, 0)
		for _, action := range tran.Actions {
			next = append(next, action.Children...)
		}
		tran.Actions = next
		*/
		fmt.Println(len(tran.Actions))
	}
}
