package gomdies

import (
	"github.com/garyburd/redigo/redis"
	"os"
	"testing"
	"time"
	"strings"
	"fmt"
	"github.com/LindenY/gomdise/mdl"
	"reflect"
)

/*
 * TODO: factor out common testing code for all packages
 */
type tsA struct {
	id string

	IntVal  int
	StrVal  string
	BytVal  byte
	BoolVal bool
}

func (ts *tsA) GetModelId() string {
	return "gomdise:test_struct:A" + ts.id
}

func (ts *tsA) SetModelId(key string) {
	lastIdx := strings.LastIndex(key, ":")
	ts.id = key[lastIdx:]
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
		"0",
		10,
		"tsA",
		0x55,
		true,
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

func TestSave(t *testing.T) {
	tsB := MakeTsB()

	gom := New(pool)
	gom.Save(tsB)
}

func TestFind(t *testing.T) {
	orig_tsB := MakeTsB()
	dest_tsB := &tsB{}

	gom := New(pool)
	key, err := gom.Save(orig_tsB)
	if err != nil {
		panic(err)
	}
	gom.Find(key, dest_tsB)

	fmt.Printf("%v \n", mdl.IfImplementsModel(reflect.TypeOf(MakeTsA())))
	fmt.Printf("%v \n", dest_tsB)
}
