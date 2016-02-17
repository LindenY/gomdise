package gomdies


import (
	"github.com/garyburd/redigo/redis"
	"fmt"
)

type Transaction struct {
	conn redis.Conn
	actions []*Action
	err error
}

type Action struct {
	name string
	args redis.Args
	handler ReplyHandler
}

type ReplyHandler func(interface{}) error

func (action *Action) String() string {
	return fmt.Sprintf("%s\t%v", action.name, action.args)
}

func NewTransaction(pool *redis.Pool) *Transaction {
	t := &Transaction{
		conn: pool.Get(),
	}
	return t
}

func (tran *Transaction) setError(err error) {
	if tran.err == nil {
		tran.err = err
	}
}

func (tran *Transaction) pushAction(action *Action) {
	if action == nil {
		return
	}
	tran.actions = append(tran.actions, action)
}

func (tran *Transaction) popAction() *Action {
	if (len(tran.actions) == 0) {
		return nil
	}
	ret := tran.actions[len(tran.actions) - 1]
	tran.actions = tran.actions[0:len(tran.actions) - 1]
	return ret
}

func (tran *Transaction) sendAction(action *Action) error {
	return tran.conn.Send(action.name, action.args)
}

func (tran *Transaction) doAction(action *Action) (interface{}, error) {
	return tran.conn.Do(action.name, action.args)
}

func (tran *Transaction) Exec() error {
	defer tran.conn.Close()

	if tran.err != nil {
		return tran.err
	}

	if len(tran.actions) == 0 {
		return nil
	}

	if len(tran.actions) == 1 {
		reply, err := tran.doAction(tran.actions[0])
		if err != nil {
			return err
		}

		if (tran.actions[0].handler != nil) {
			return tran.actions[0].handler(reply)
		}
	}

	if err := tran.conn.Send("MULTI"); err != nil {
		return err
	}

	for _, a := range tran.actions {
		if err := tran.sendAction(a); err != nil {
			return err
		}
 	}

	replies, err := redis.Values(tran.conn.Do("EXEC"))
	if err != nil {
		return err
	}

	for i, reply := range replies {
		if tran.actions[i].handler != nil {
			return tran.actions[i].handler(reply)
		}
	}

	return nil
}