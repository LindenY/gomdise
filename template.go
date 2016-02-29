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

func (pt *proxyTemplate) handle(tran *Transaction, reply interface{}) error {
	pt.wg.Wait()
	return pt.dest.handle(tran, reply)
}

func (pt *proxyTemplate) engrave(tran *Transaction, args interface{}) error {
	pt.wg.Wait()
	return pt.dest.engrave(tran, args)
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

func newArrayFindTemplate(t reflect.Type) *arrayFindTemplate {
	aft := &arrayFindTemplate{
		elemEgr: findTemplateForType(t.Elem()).engrave,
	}
	return aft
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

func (pft *pointerFindTemplate) handle(tran *Transaction, reply interface{}) error {
	return pft.elemTpl.handle(tran, reply)
}

func (pft *pointerFindTemplate) engrave(tran *Transaction, args interface{}) error {
	return pft.elemTpl.engrave(tran, args)
}

func newPointerFindTemplate(t reflect.Type) ActionTemplate {
	pft := &pointerFindTemplate{findTemplateForType(t.Elem())}
	return pft
}

type voidFindTemplate struct {
}

func (vft *voidFindTemplate) engrave(tran *Transaction, args interface{}) error {
	return nil
}

func (vft *voidFindTemplate) handle(tran *Transaction, reply interface{}) error {
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
	return fmt.Sprint("[%v]\t%s", uaerr.Args, uaerr.Msg)
}
