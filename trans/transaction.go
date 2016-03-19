package trans

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
)

type Transaction struct {
	conn    redis.Conn
	Actions []*Action
	State   interface{}
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

func (tran *Transaction) sendAction(action *Action) error {
	log.Printf("Transaction: sending action %v \n", action)
	switch action.Type {
	case CmdAction:
		return tran.conn.Send(action.Name, action.Args...)
	case ScriptAction:
		return action.Script.Send(tran.conn, action.Args...)
	default:
		return errors.New(fmt.Sprintf("Unsupported Action Type %v", action.Type))
	}
}

func (tran *Transaction) doAction(action *Action) (interface{}, error) {
	log.Printf("Transaction: doing action %v \n", action)
	switch action.Type {
	case CmdAction:
		return tran.conn.Do(action.Name, action.Args...)
	case ScriptAction:
		return action.Script.Do(tran.conn, action.Args...)
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported Action Type %v", action.Type))
	}
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

	prev := tran.Actions
	tran.Actions = make([]*Action, 0)
	for i, action := range prev {
		log.Printf("Transaction: receive reply [%v] for action [%v]\n", replies[i], action)

		action.Reply = replies[i]
		action.handle(tran, replies[i])
	}
}

func (tran *Transaction) Exec() interface{} {
	defer tran.conn.Close()
	for len(tran.Actions) > 0 {
		log.Printf("Transaction: >>>\t executing transaction with %d actions \n", len(tran.Actions))
		tran.exec()
		log.Printf("Transaction: >>>\t executed transaction")
	}
	return tran.State
}
