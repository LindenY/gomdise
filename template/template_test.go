package gomdies

import (
	"reflect"
	"testing"
	"github.com/LindenY/gomdise/trans"
	"github.com/garyburd/redigo/redis"
	"time"
	"os"
	"fmt"
)

type tsA struct {
	IntVal int
	StrVal string
	BytVal byte
}

type tsB struct {
	MStrSrt map[string]tsA
	LSrtVal []tsA
}

type tsC struct {
	MStrStrSrt map[string]map[string]*tsA
	LLSrtVal   [][]*tsA
}

type tsD struct {
	tsB
	tsC
}

type tsE struct {
	TsB tsB
	TsC tsC
}

func MakeTsA() *tsA {
	return &tsA{
		10,
		"tsA",
		0x55,
	}
}

func MakeTsB() *tsB {
	tsB := &tsB{
		make(map[string]tsA),
		make([]tsA, 2, 2),
	}
	tsB.MStrSrt["TsA0"] = *MakeTsA()
	tsB.MStrSrt["TsA1"] = *MakeTsA()
	tsB.LSrtVal[0] = *MakeTsA()
	tsB.LSrtVal[1] = *MakeTsA()
	return tsB
}

func MakeTsC() *tsC {
	tsC := &tsC{
		make(map[string]map[string]*tsA),
		make([][]*tsA, 2),
	}
	tsC.MStrStrSrt["map0"] = make(map[string]*tsA)
	tsC.MStrStrSrt["map0"]["tsA"] = MakeTsA()
	tsC.MStrStrSrt["map1"] = make(map[string]*tsA)
	tsC.MStrStrSrt["map1"]["tsA"] = MakeTsA()

	tsC.LLSrtVal[0] = make([]*tsA, 2)
	tsC.LLSrtVal[0][0] = MakeTsA()
	tsC.LLSrtVal[0][1] = MakeTsA()
	tsC.LLSrtVal[1] = make([]*tsA, 2)
	tsC.LLSrtVal[1][0] = MakeTsA()
	tsC.LLSrtVal[1][1] = MakeTsA()
	return tsC
}

var pool *redis.Pool

func TestMain(m *testing.M) {
	pool = &redis.Pool{
		MaxIdle:     1,
		IdleTimeout: 3 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "184.107.247.74:16379")
			if err != nil {
				return nil, err
			}

			if password := ""; len(password) > 0 {
				if _, err := conn.Do("AUTH", password); err != nil {
					conn.Close()
					return nil, err
				}
			}

			return conn, err
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
	os.Exit(m.Run())
}

func TestFindTemplateTsB(t *testing.T) {
	tpl := tcache_find.GetTemplate(reflect.TypeOf(tsB{}))
	tran := trans.NewTransaction(pool)
	tpl.engrave(&tran.Actions, "gomdies.tsB:0b4063db-81ae-46cc-99e3-b64863caf0ce")
	fmt.Println(tran.Actions)
	tran.Exec()
}
