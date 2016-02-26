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

type ActionTemplate struct {
	Name     string
	Handler  ReplyHandler
	Engraver ArgsEngraver
}

type Action struct {
	ActionTemplate
	Args redis.Args
}

type ReplyHandler func(trans *Transaction, reply interface{}) error

type ArgsEngraver func(trans *Transaction, args interface{}) error

func (atpl *ActionTemplate) handle(tran *Transaction, reply interface{}) error {
	if atpl.Handler != nil {
		return atpl.Handler(tran, reply)
	}
	return nil
}

func (atpl *ActionTemplate) engrave(tran *Transaction, args interface{}) error {
	if atpl.Handler != nil {
		return atpl.Engraver(tran, args)
	}
	return nil
}

func (action *Action) String() string {
	return fmt.Sprintf("%s\t%v", action.ActionTemplate.Name, action.Args)
}

func NewTransaction(pool *redis.Pool) *Transaction {
	t := &Transaction{
		conn: pool.Get(),
	}
	return t
}

func (tran *Transaction) setError(err error) {
	if tran.Err == nil {
		tran.Err = err
	}
}

func (tran *Transaction) sendAction(action *Action) error {
	return tran.conn.Send(action.ActionTemplate.Name, action.Args)
}

func (tran *Transaction) doAction(action *Action) (interface{}, error) {
	return tran.conn.Do(action.ActionTemplate.Name, action.Args)
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
