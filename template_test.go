package gomdies

import (
	_ "github.com/garyburd/redigo/redis"
	"testing"
	_ "time"
	_ "os"
	"reflect"
)

/*
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
*/


func TestFindTemplateTsB(t *testing.T) {

	tpl := findTemplateForType(reflect.TypeOf(tsB{}))

	tran := NewTransaction(pool)
	tpl.engrave(tran, "gomdies.tsB:0b4063db-81ae-46cc-99e3-b64863caf0ce")
	printActions(tran.Actions)
	if err := tran.Exec(); err != nil {
		panic(err)
	}
}