package gomdies

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
	"errors"
	"sync"
)

type parseState struct {
	actions []*Action
	target *interface{}
}


func (pstate *parseState)pushAction(action *Action) {
	pstate.actions = append(pstate.actions, action)
}

func (pstate *parseState)popAction() *Action {
	if pstate.actions == nil || len(pstate.actions) == 0 {
		return nil
	}

	ret := pstate.actions[len(pstate.actions)-1]
	pstate.actions = pstate.actions[0:len(pstate.actions)-1]
	return ret
}


// Save Parse Functions
type parseFunc func(pstate *parseState, v reflect.Value)

var parserCache struct {
	sync.RWMutex
	m map[reflect.Type]parseFunc
}

func saveParser(t reflect.Type) parseFunc {
	parserCache.RLock()
	f := parserCache.m[t]
	parserCache.RUnlock()
	if f != nil {
		return f
	}

	parserCache.Lock()
	if parserCache.m == nil {
		parserCache.m = make(map[reflect.Type]parseFunc)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	parserCache.m[t] = func(pstate *parseState, v reflect.Value) {
		wg.Wait()
		f(pstate, v)
	}
	parserCache.Unlock()

	f = newSaveParser(t)
	wg.Done()
	parserCache.Lock()
	parserCache.m[t] = f
	parserCache.Unlock()

	return f
}

func newSaveParser(t reflect.Type) parseFunc {

	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return primitiveSaveParser
	case reflect.Array, reflect.Slice:
		return newArraySaveParser(t)
	case reflect.Map:
		return newMapSaveParser(t)
	case reflect.Struct:
		return newStructSaveParser(t)
	case reflect.Ptr:
		return newPointerSaveParser(t)
	default:
		return unsupportedTypeParser
	}
}


func primitiveSaveParser(pstate *parseState, v reflect.Value) {
	if len(pstate.actions) == 0 {
		panic(errors.New("Primitive type is not supported"))
	}

	curr := pstate.actions[len(pstate.actions)-1]
	curr.args = curr.args.Add(v.Interface())
}


type arraySaveParser struct {
	elemFunc parseFunc
}

func (asp *arraySaveParser) parse(pstate *parseState, v reflect.Value) {
	arrKey := newKey(v)
	prev := pstate.popAction()
	curr := &Action{
		name: "RPUSH",
		args:redis.Args{arrKey},
	}
	pstate.pushAction(curr)

	n := v.Len()
	for i := 0; i<n; i++ {
		asp.elemFunc(pstate, v)
	}

	if prev != nil {
		prev.args = prev.args.Add(arrKey)
		pstate.pushAction(prev)
	}
}

func newArraySaveParser(t reflect.Type) parseFunc {
	asp := &arraySaveParser{saveParser(t.Elem())}
	return asp.parse
}


type mapSaveParser struct {
	elemFunc parseFunc
}

func (msp *mapSaveParser) parse(pstate *parseState, v reflect.Value) {
	mapKey := newKey(v)
	prev := pstate.popAction()
	curr := &Action{
		name:"HSET",
		args:redis.Args{mapKey},
	}
	pstate.pushAction(curr)

	skeys := v.MapKeys()
	for _, skey := range skeys {
		curr.args = curr.args.Add(skey)
		msp.elemFunc(pstate, v)
	}

	if prev != nil {
		prev.args = prev.args.Add(mapKey)
		pstate.pushAction(prev)
	}
}

func newMapSaveParser(t reflect.Type) parseFunc {
	if t.Key().Kind() != reflect.String {
		return unsupportedTypeParser
	}

	msp := &mapSaveParser{saveParser(t.Elem())}
	return msp.parse
}


type structSaveParser struct {
	spec *structSpec
	elemFuncs []parseFunc
}

func (ssp *structSaveParser) parse(pstate *parseState, v reflect.Value) {
	srtKey := newKey(v)
	prev := pstate.popAction()
	curr := &Action{
		name:"HSET",
		args:redis.Args{srtKey},
	}
	pstate.pushAction(curr)

	for i, fldSpec := range ssp.spec.fields {
		curr.args = curr.args.Add(fldSpec.name)
		ssp.elemFuncs[i](pstate, v)
	}

	if prev != nil {
		prev.args = prev.args.Add(srtKey)
		pstate.pushAction(prev)
	}
}

func newStructSaveParser(t reflect.Type) parseFunc {
	srtSpec := structSpecForType(t)
	ssp := &structSaveParser{
		spec: srtSpec,
		elemFuncs:make([]parseFunc, len(srtSpec.fields)),
	}

	for i, fldSpec := range ssp.spec.fields {
		ssp.elemFuncs[i] = saveParser(fldSpec.typ)
	}
	return ssp.parse
}


type pointerSaveParser struct {
	elemFunc parseFunc
}

func (psp *pointerSaveParser) parse(pstate *parseState, v reflect.Value) {
	if v.IsNil() {
		curr := pstate.actions[len(pstate.actions)-1]
		curr.args = curr.args.Add("NULL")
		return
	}
	psp.elemFunc(pstate, v.Elem())
}

func newPointerSaveParser(t reflect.Type) parseFunc {
	psp := &pointerSaveParser{saveParser(t.Elem())}
	return psp.parse
}


func unsupportedTypeParser(pstate *parseState, v reflect.Value) {
	panic(&UnsupportedTypeError{v.Type()})
}

type UnsupportedTypeError struct {
	Type reflect.Type
}