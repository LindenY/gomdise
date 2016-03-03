package gomdies

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"sync"
)

type ActionTemplate interface {
	handle(tran *Transaction, action *Action, reply interface{}) error
	engrave(args ...interface{}) *Action
}

type ArgsEngraver func(args ...interface{}) *Action

var findTemplateCache struct {
	sync.RWMutex
	m map[reflect.Type]ActionTemplate
}

func findTemplateForType(t reflect.Type) ActionTemplate {
	findTemplateCache.RLock()
	tpl := findTemplateCache.m[t]
	findTemplateCache.RUnlock()
	if tpl != nil {
		return tpl
	}

	findTemplateCache.Lock()
	if findTemplateCache.m == nil {
		findTemplateCache.m = make(map[reflect.Type]ActionTemplate)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	findTemplateCache.m[t] = &proxyTemplate{wg, tpl}
	findTemplateCache.Unlock()

	tpl = newFindTemplateForType(t)
	wg.Done()
	findTemplateCache.Lock()
	findTemplateCache.m[t] = tpl
	findTemplateCache.Unlock()
	return tpl
}

type proxyTemplate struct {
	wg   sync.WaitGroup
	dest ActionTemplate
}

func (pt *proxyTemplate) handle(tran *Transaction, action *Action, reply interface{}) error {
	pt.wg.Wait()
	return pt.dest.handle(tran, action, reply)
}

func (pt *proxyTemplate) engrave(args ...interface{}) *Action {
	pt.wg.Wait()
	return pt.dest.engrave(args...)
}

func newFindTemplateForType(t reflect.Type) ActionTemplate {
	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return newVoidFindTemplate(t)
	case reflect.Array, reflect.Slice:
		return newArrayFindTemplate(t)
	case reflect.Map:
		return newMapFindTemplate(t)
	case reflect.Struct:
		return newStructFindTemplate(t)
	case reflect.Ptr:
		return newPointerFindTemplate(t)
	default:
		return nil
	}
}

type arrayFindTemplate struct {
	elemEgr ArgsEngraver
}

func (aft *arrayFindTemplate) handle(tran *Transaction, action *Action, reply interface{}) error {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		return err
	}
	for _, rpy := range replies {
		action.addChild(aft.elemEgr(rpy))
	}
	return nil
}

func (aft *arrayFindTemplate) engrave(args ...interface{}) *Action {
	key, ok := args[0].(string)
	if !ok {
		_, err := redis.Scan(args, &key)
		if err != nil {
			panic(err)
		}
	}

	action := &Action{
		Name:    "LRANGE",
		Args:    redis.Args{key, 0, -1},
		Handler: aft.handle,
	}
	return action
}

func newArrayFindTemplate(t reflect.Type) *arrayFindTemplate {
	aft := &arrayFindTemplate{
		elemEgr: findTemplateForType(t.Elem()).engrave,
	}
	return aft
}

type mapFindTemplate struct {
	elemEgr ArgsEngraver
}

func (mft *mapFindTemplate) handle(tran *Transaction, action *Action, reply interface{}) error {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		return err
	}
	toggle := false
	for _, rpy := range replies {
		toggle = !toggle
		if toggle {
			continue
		}
		action.addChild(mft.elemEgr(rpy))
	}
	return nil
}

func (mft *mapFindTemplate) engrave(args ...interface{}) *Action {
	key, ok := args[0].(string)
	if !ok {
		_, err := redis.Scan(args, &key)
		if err != nil {
			panic(err)
		}
	}

	action := &Action{
		Name:    "HGETALL",
		Args:    redis.Args{key},
		Handler: mft.handle,
	}
	return action
}

func newMapFindTemplate(t reflect.Type) *mapFindTemplate {
	mft := &mapFindTemplate{
		elemEgr: findTemplateForType(t.Elem()).engrave,
	}
	return mft
}

type structFindTemplate struct {
	spec     *structSpec
	elemEgrs []ArgsEngraver
}

func (sft *structFindTemplate) handle(tran *Transaction, action *Action, reply interface{}) error {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		return err
	}
	toggle := false
	for i, rpy := range replies {
		toggle = !toggle
		if toggle {
			continue
		}
		action.addChild(sft.elemEgrs[(i-1)/2](rpy))
	}
	return nil
}

func (sft *structFindTemplate) engrave(args ...interface{}) *Action {
	key, ok := args[0].(string)
	if !ok {
		_, err := redis.Scan(args, &key)
		if err != nil {
			panic(err)
		}
	}

	action := &Action{
		Name:    "HGETALL",
		Args:    redis.Args{key},
		Handler: sft.handle,
	}
	return action
}

func newStructFindTemplate(t reflect.Type) *structFindTemplate {
	srtSpec := structSpecForType(t)
	sft := &structFindTemplate{
		spec:     srtSpec,
		elemEgrs: make([]ArgsEngraver, len(srtSpec.fields)),
	}

	for i, fldSpec := range sft.spec.fields {
		sft.elemEgrs[i] = findTemplateForType(fldSpec.typ).engrave
	}
	return sft
}

type pointerFindTemplate struct {
	elemTpl ActionTemplate
}

func (pft *pointerFindTemplate) handle(tran *Transaction, action *Action, reply interface{}) error {
	return pft.elemTpl.handle(tran, action, reply)
}

func (pft *pointerFindTemplate) engrave(args ...interface{}) *Action {
	return pft.elemTpl.engrave(args)
}

func newPointerFindTemplate(t reflect.Type) ActionTemplate {
	pft := &pointerFindTemplate{findTemplateForType(t.Elem())}
	return pft
}

type voidFindTemplate struct {
}

func (vft *voidFindTemplate) handle(tran *Transaction, action *Action, reply interface{}) error {
	return nil
}

func (vft *voidFindTemplate) engrave(args ...interface{}) *Action {
	return nil
}

var vft *voidFindTemplate

func newVoidFindTemplate(t reflect.Type) *voidFindTemplate {
	if vft == nil {
		vft = &voidFindTemplate{}
	}
	return vft
}

type UnsupportArgsError struct {
	Msg  string
	Args interface{}
}

func (uaerr UnsupportArgsError) Error() string {
	return fmt.Sprintf("[%v]\t%s", uaerr.Args, uaerr.Msg)
}
