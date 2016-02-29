package gomdies

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type Transaction struct {
	conn    redis.Conn
	offset  int
	Actions []*Action
	Err     error
}

type Action struct {
	Name    string
	Args    redis.Args
	Handler ReplyHandler
}

type ReplyHandler func(tran *Transaction, reply interface{}) error

func (action *Action) handle(trans *Transaction, reply interface{}) error {
	if action.Handler != nil {
		return action.Handler(trans, reply)
	}
	return nil
}

func (action *Action) String() string {
	return fmt.Sprintf("%s\t%v", action.Name, action.Args)
}

func NewTransaction(pool *redis.Pool) *Transaction {
	t := &Transaction{
		conn: pool.Get(),
	}
	return t
}

func (tran *Transaction) pushAction(action *Action) {
	if action == nil {
		return
	}
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
	return tran.conn.Send(action.Name, action.Args)
}

func (tran *Transaction) doAction(action *Action) (interface{}, error) {
	return tran.conn.Do(action.Name, action.Args)
}

func (tran *Transaction) numUnexecActions() int {
	return len(tran.Actions) - tran.offset - 1
}

func (tran *Transaction) exec() error {
	if tran.numUnexecActions() == 1 {
		reply, err := tran.doAction(tran.Actions[0])
		if err != nil {
			return err
		}
		return tran.Actions[0].handle(reply)
	}

	if err := tran.conn.Send("MULTI"); err != nil {
		return err
	}
	for i := tran.offset; i < len(tran.Actions); i++ {
		if err := tran.sendAction(tran.Actions[i]); err != nil {
			return err
		}
	}
	replies, err := redis.Values(tran.conn.Do("EXEC"))
	if err != nil {
		return err
	}
	for i, reply := range replies {
		if err := tran.Actions[tran.offset+i].handle(reply); err != nil {
			break
		}
	}
	tran.offset = len(tran.Actions)
	return err
}

func (tran *Transaction) Exec() error {
	defer tran.conn.Close()
	if tran.Err != nil {
		return tran.Err
	}
	for tran.numUnexecActions() > 0 {
		if err := tran.exec(); err != nil {
			return err
		}
	}
	return nil
}
