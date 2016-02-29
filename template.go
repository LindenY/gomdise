package gomdies

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"sync"
)

type ActionTemplate interface {
	handle(tran *Transaction, reply interface{}) error
	engrave(tran *Transaction, args interface{}) error
}

type ArgsEngraver func(tran *Transaction, args interface{}) error


var findTemplateCache struct {
	sync.RWMutex
	m map[reflect.Type]ArgsEngraver
}

func findTemplateForType(t reflect.Type) ArgsEngraver {
	findTemplateCache.RLock()
	tpl := findTemplateCache.m[t]
	findTemplateCache.RUnlock()
	if tpl != nil {
		return tpl
	}

	findTemplateCache.Lock()
	if findTemplateCache.m == nil {
		findTemplateCache.m = make(map[reflect.Type]ArgsEngraver)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	findTemplateCache.m[t] = func(tran *Transaction, reply interface{}) error {
		wg.Wait()
		tpl(tran, reply)
		return nil
	}
	findTemplateCache.Unlock()

	tpl = newFindTemplateForType(t)
	wg.Done()
	findTemplateCache.Lock()
	findTemplateCache.m[t] = tpl
	findTemplateCache.Unlock()
	return tpl
}

func newFindTemplateForType(t reflect.Type) ArgsEngraver {
	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return voidFindArgsEngraver
	case reflect.Array, reflect.Slice:
		return newArrayFindTemplate(t)
	case reflect.Map:
		return newMapFindTemplate(t)
	case reflect.Struct:
		return newStructFindTemplate(t)
	case reflect.Ptr:
		return newPointerSaveParser(t)
	default:
		return nil
	}
}

type arrayFindTemplate struct {
	elemEgr ArgsEngraver
}

func (aft *arrayFindTemplate) handle(tran *Transaction, reply interface{}) error {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		return err
	}
	for _, rpy := range replies {
		err = aft.elemEgr(tran, rpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (aft *arrayFindTemplate) engrave(tran *Transaction, args interface{}) error {
	if _, ok := args.(string); !ok {
		return UnsupportArgsError{
			Msg:  "Unsupported args for arrayFindTemplate.engrave",
			Args: args,
		}
	}

	action := &Action{
		Name:    "LRANGE",
		Args:    redis.Args{args, 0, -1},
		Handler: aft.handle,
	}
	tran.pushAction(action)
	return nil
}

func newArrayFindTemplate(t reflect.Type) ArgsEngraver {
	aft := arrayFindTemplate{
		elemEgr: findTemplateForType(t.Elem()),
	}
	return aft.engrave
}

type mapFindTemplate struct {
	elemEgr ArgsEngraver
}

func (mft *mapFindTemplate) handle(tran *Transaction, reply interface{}) error {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		return err
	}
	toggle := true
	for _, rpy := range replies {
		if toggle {
			toggle = false
			continue
		}
		err = mft.elemEgr(tran, rpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mft *mapFindTemplate) engrave(tran *Transaction, args interface{}) error {
	if _, ok := args.(string); !ok {
		return UnsupportArgsError{
			Msg:  "Unsupported args for mapFindTemplate.engrave",
			Args: args,
		}
	}

	action := &Action{
		Name:    "HGETALL",
		Args:    redis.Args{args},
		Handler: mft.handle,
	}
	tran.pushAction(action)
	return nil
}

func newMapFindTemplate(t reflect.Type) ArgsEngraver {
	mft := mapFindTemplate{
		elemEgr: findTemplateForType(t.Elem()),
	}
	return mft.engrave
}

type structFindTemplate struct {
	spec     *structSpec
	elemEgrs []ArgsEngraver
}

func (sft *structFindTemplate) handle(tran *Transaction, reply interface{}) error {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		return err
	}
	toggle := true
	for i, rpy := range replies {
		if toggle {
			toggle = false
			continue
		}
		err = sft.elemEgrs[(i-1)/2](tran, rpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sft *structFindTemplate) engrave(tran *Transaction, args interface{}) error {
	if _, ok := args.(string); !ok {
		return UnsupportArgsError{
			Msg:  "Unsupported args for structFindTemplate.engrave",
			Args: args,
		}
	}

	action := &Action{
		Name:    "HGETALL",
		Args:    redis.Args{args},
		Handler: sft.handle,
	}
	tran.pushAction(action)
	return nil
}

func newStructFindTemplate(t reflect.Type) ArgsEngraver {
	srtSpec := structSpecForType(t)
	sft := &structFindTemplate{
		spec:     srtSpec,
		elemEgrs: make([]ArgsEngraver, len(srtSpec.fields)),
	}

	for i, fldSpec := range sft.spec.fields {
		sft.elemEgrs[i] = findTemplateForType(fldSpec.typ)
	}
	return sft.engrave
}

type pointerFindTemplate struct {
	elemEgr ArgsEngraver
}

func (pft *pointerFindTemplate) handle(tran *Transaction, reply interface{}) error {
	return nil
}

func (pft *pointerFindTemplate) engrave(tran *Transaction, args interface{}) error {
	return  nil
}

func newPointerFindTemplate(t reflect.Type) ArgsEngraver {
	pft := &pointerFindTemplate{findTemplateForType(t.Elem())}
	return pft.engrave
}

func voidFindArgsEngraver(tran *Transaction, args interface{}) error {
	return nil
}

type UnsupportArgsError struct {
	Msg  string
	Args interface{}
}

func (uaerr *UnsupportArgsError) Error() string {
	return fmt.Sprint("[%v]\t%s", uaerr.Args, uaerr.Msg)
}
