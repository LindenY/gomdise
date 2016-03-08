package gomdies

import (
	"github.com/garyburd/redigo/redis"
	"sync"
	"runtime"
	"errors"
	"reflect"
)

type Gomdise struct {
	sync.RWMutex
	pool redis.Pool
	opts map[string]interface{}
}

func (gom *Gomdise) SetOption(key string, val interface{}) {
	gom.Lock()
	defer gom.Unlock()
	gom.opts[key] = val
}

func (gom *Gomdise) GetOption(key string) interface{} {
	gom.RLock()
	defer gom.RUnlock()
	return gom.opts[key]
}


func (gom *Gomdise) Save(args interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				err = errors.New(s)
			}
			err = r.(error)
		}
	}()


	return nil
}

func (gom *Gomdise) Find(key string, dest interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				err = errors.New(s)
			}
			err = r.(error)
		}
	}()

	tpl := findTemplateForType(reflect.TypeOf(dest))
	root := tpl.engrave(key)
	tran := NewTransaction(gom.pool)
	tran.Actions = append(tran.Actions, root)
	tran.Exec()
	Decode(root, dest)
	return nil
}


func Init(pool redis.Pool) *Gomdise {
	gom := &Gomdise{
		pool:pool,
		opts:make(map[string]interface{}, 0),
	}
	return gom
}