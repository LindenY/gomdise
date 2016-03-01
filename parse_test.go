package gomdies

import (
	"testing"
	"fmt"
	"time"
	"github.com/garyburd/redigo/redis"
	"os"
)


var pool *redis.Pool

func TestMain(m *testing.M) {
	pool = &redis.Pool {
		MaxIdle : 1,
		IdleTimeout : 3 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "184.107.247.74:16379")
			if err != nil {
				return nil, err
			}

			if password := ""; len(password) > 0 {
				if _, err := conn.Do("AUTH", password); err != nil {
					conn.Close();
					return nil, err
				}
			}

			return conn, err;
		},
		TestOnBorrow : func (conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
	os.Exit(m.Run())
}

func TestParseTsB(t *testing.T) {
	tsB := MakeTsB()

	actions, err := parseSave(tsB)
	if err != nil {
		panic(err)
	}
	printActions(actions)

	tran := NewTransaction(pool)
	tran.Actions = actions
	if err := tran.Exec(); err != nil {
		panic(err)
	}
}

func TestParseTsC(t *testing.T) {
	tsC := MakeTsC()

	actions, err := parseSave(tsC)
	if err != nil {
		panic(err)
	}
	printActions(actions)

	tran := NewTransaction(pool)
	tran.Actions = actions
	if err := tran.Exec(); err != nil {
		panic(err)
	}
}


func printActions(actions  []*Action) {
	fmt.Printf("Num of actions: %d\n", len(actions))
	for i, a := range actions {
		fmt.Printf("\t[%d]:\t%v\n", i, a)
	}
}
