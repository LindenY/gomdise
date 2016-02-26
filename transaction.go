package gomdies

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type Chain struct {
	trans []*Transaction
	err   error
}

type Transaction struct {
	conn    redis.Conn
	actions []*Action
	err     error
}

type Action struct {
	name    string
	args    redis.Args

	handler ReplyHandler
	injector ArgsInjector
}

type ReplyHandler func(interface{}) error

type ArgsInjector func(...interface{}) error


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

func (tran *Transaction) sendAction(action *Action) error {
	return tran.conn.Send(action.name, action.args)
}

func (tran *Transaction) doAction(action *Action) (interface{}, error) {
	return tran.conn.Do(action.name, action.args)
}

func (tran *Transaction) ExecWithHandler() error {
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

		if tran.actions[0].handler != nil {
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

func (tran *Transaction) Exec(conn redis.Conn) ([]interface{}, error) {
	if tran.err != nil {
		return nil, tran.err
	}
	if len(tran.actions) == 0 {
		return nil, nil
	}

	tran.conn = conn
	if len(tran.actions) == 1 {
		reply, err := tran.doAction(tran.actions[0])
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
	if err := tran.conn.Send("MULTI"); err != nil {
		return nil, err
	}
	for _, a := range tran.actions {
		if err := tran.sendAction(a); err != nil {
			return nil, err
		}
	}
	return redis.Values(conn.Do("EXEC"))
}



func (tran *Transaction) InjectArgs(args ...interface{}) error {
	argsIdx := 0;
	for _, action := range tran.actions {
		if action.injector == nil {
			continue
		}

		err := action.injector(args[argsIdx])
		if err != nil {
			return err
		}
	}
	return nil
}


func (chain *Chain)setError(err error) {
	if chain.err != nil {
		chain.err = err
	}
}

func (chain *Chain)Exec(conn redis.Conn) ([]interface{}, error) {
	defer conn.Close()

	replies := make([]interface{}, len(chain.trans))
	for i, tran := range chain.trans {
		reply, err := tran.Exec(conn)
		if err != nil {
			return nil, err
		}

		replies = append(replies, reply)
		if i < len(chain.trans) {
			err := chain.trans[i+1].InjectArgs(reply)
			if err != nil {
				return nil, err
			}
		}
	}
	return replies, nil
}
