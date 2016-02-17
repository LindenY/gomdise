package gomdies


import (
	"github.com/garyburd/redigo/redis"
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

func (tran *Transaction) Command(name string, args redis.Args, handler ReplyHandler) {
	tran.actions = append(tran.actions, &Action{
		name: name,
		args: args,
		handler: handler,
	})
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