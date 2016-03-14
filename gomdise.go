package gomdies

import (
	"errors"
	"github.com/LindenY/gomdies/mdl"
	"github.com/LindenY/gomdies/tpl"
	"github.com/LindenY/gomdies/trans"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"runtime"
	"sync"
)

type Gomdise struct {
	sync.RWMutex
	pool *redis.Pool
	opts map[string]interface{}
}

func (gom *Gomdise) SetOption(key string, val interface{}) {
	gom.Lock()
	defer gom.Unlock()
	gom.opts[key] = val
}

func (gom *Gomdise) GetOption(key string) (interface{}, bool) {
	gom.RLock()
	defer gom.RUnlock()
	val, ok := gom.opts[key]
	return val, ok
}

func (gom *Gomdise) Save(arg interface{}) (key string, err error) {
	defer func() {
		err = errorRecover()
		if err != nil {
			key = ""
		}
	}()

	tpl := tpl.TCSave.GetTemplate(reflect.TypeOf(arg))
	tran := trans.NewTransaction(gom.pool)
	v := reflect.ValueOf(arg)
	key = mdl.NewKey(v)
	tpl.Engrave(&tran.Actions, key, v)
	tran.Exec()
	return key, nil
}

func (gom *Gomdise) SaveWithKey(arg interface{}, key string) (err error) {
	defer func() {
		err = errorRecover()
	}()
	tpl := tpl.TCSave.GetTemplate(reflect.TypeOf(arg))
	tran := trans.NewTransaction(gom.pool)
	v := reflect.ValueOf(arg)
	tpl.Engrave(&tran.Actions, key, v)
	tran.Exec()
	return nil
}

func (gom *Gomdise) Find(key string, dest interface{}) (err error) {
	defer func() {
		err = errorRecover()
	}()

	tpl := tpl.TCFind.GetTemplate(reflect.TypeOf(dest))
	tran := trans.NewTransaction(gom.pool)
	tpl.Engrave(&tran.Actions, key)
	root := tran.Actions[0]
	tran.Exec()
	mdl.Decode(root, dest)
	return nil
}

func New(pool *redis.Pool) *Gomdise {
	gom := &Gomdise{
		pool: pool,
		opts: make(map[string]interface{}, 0),
	}
	return gom
}

func errorRecover() (err error) {
	if r := recover(); r != nil {
		if _, ok := r.(runtime.Error); ok {
			panic(r)
		}
		if s, ok := r.(string); ok {
			err = errors.New(s)
		}
		err = r.(error)
	}
	return nil
}
