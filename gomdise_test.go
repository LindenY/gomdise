package gomdise

import (
	"github.com/garyburd/redigo/redis"
	"os"
	"testing"
	"time"
	"strings"
	"log"
	"io/ioutil"
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
	lastIdx := strings.LastIndex(key, "A")
	ts.id = key[lastIdx+1:]
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

func TestSaveTsB(t *testing.T) {
	tsB := MakeTsB()
	gom := New(pool)
	gom.Save(tsB)
}

func TestSaveTsC(t *testing.T) {
	tsC := MakeTsC()
	gom := New(pool)
	gom.Save(tsC)
}

func TestSaveWithKeyForMapWithInterfaceValue(t *testing.T) {
	m := make(map[string]interface{})
	m["str"] = "key0"
	m["srt"] = MakeTsA()
	gom := New(pool)
	gom.SaveWithKey(m, "TestSaveWithKeyForMapWithInterfaceValue")
}

func TestFindTsB(t *testing.T) {
	orig_tsB := MakeTsB()
	dest_tsB := &tsB{}

	gom := New(pool)
	key, err := gom.Save(orig_tsB)
	if err != nil {
		panic(err)
	}
	gom.Find(key, dest_tsB)
}

func TestFindTsC(t *testing.T) {
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	orig_tsC := MakeTsC()
	dest_tsC := &tsC{}

	gom := New(pool)
	key, err := gom.Save(orig_tsC)
	if err != nil {
		panic(err)
	}
	gom.Find(key, dest_tsC)
}

func TestFindWithKeyForMapWithInterfaceValue(t *testing.T) {
	m := make(map[string]interface{})
	m["str"] = "key0"
	m["srt"] = MakeTsA()
	gom := New(pool)
	gom.SaveWithKey(m, "TestSaveWithKeyForMapWithInterfaceValue")

	gom.Find("TestSaveWithKeyForMapWithInterfaceValue", make(map[string]interface{}))
}
