package gomdies

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"runtime"
	"sync"
)

type parseState struct {
	actions []*Action
	target  interface{}
}

func (pstate *parseState) pushAction(action *Action) {
	pstate.actions = append(pstate.actions, action)
}

func (pstate *parseState) popAction() *Action {
	if pstate.actions == nil || len(pstate.actions) == 0 {
		return nil
	}

	ret := pstate.actions[len(pstate.actions)-1]
	pstate.actions = pstate.actions[0 : len(pstate.actions)-1]
	return ret
}

func (pstate *parseState) marshal(v interface{}, prsFunc parseFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()
	prsFunc(pstate, reflect.ValueOf(v))
	return nil
}

type parseFunc func(pstate *parseState, v reflect.Value)

// Save Parse Functions

func parseSave(v interface{}) (actions []*Action, err error) {
	pstate := &parseState{
		actions: make([]*Action, 0),
		target:  v,
	}
	err = pstate.marshal(v, saveParser(reflect.TypeOf(v)))
	if err != nil {
		return nil, err
	}
	return pstate.actions, nil
}

var saveParserCache struct {
	sync.RWMutex
	m map[reflect.Type]parseFunc
}

func saveParser(t reflect.Type) parseFunc {

	fmt.Printf("SaveParser: %v \n", t)

	saveParserCache.RLock()
	f := saveParserCache.m[t]
	saveParserCache.RUnlock()
	if f != nil {
		return f
	}

	saveParserCache.Lock()
	if saveParserCache.m == nil {
		saveParserCache.m = make(map[reflect.Type]parseFunc)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	saveParserCache.m[t] = func(pstate *parseState, v reflect.Value) {
		wg.Wait()
		f(pstate, v)
	}
	saveParserCache.Unlock()

	f = newSaveParser(t)
	wg.Done()
	saveParserCache.Lock()
	saveParserCache.m[t] = f
	saveParserCache.Unlock()

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
	curr.Args = curr.Args.Add(v.Interface())
}

type arraySaveParser struct {
	elemFunc parseFunc
}

func (asp *arraySaveParser) parse(pstate *parseState, v reflect.Value) {
	arrKey := newKey(v)
	prev := pstate.popAction()
	curr := &Action{
		Name: "RPUSH",
		Args: redis.Args{arrKey},
	}
	pstate.pushAction(curr)

	n := v.Len()
	for i := 0; i < n; i++ {
		asp.elemFunc(pstate, v.Index(i))
	}

	if prev != nil {
		prev.Args = prev.Args.Add(arrKey)
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
		Name: "HSET",
		Args: redis.Args{mapKey},
	}
	pstate.pushAction(curr)

	skeys := v.MapKeys()
	for _, skey := range skeys {
		curr.Args = curr.Args.Add(skey)
		msp.elemFunc(pstate, v.MapIndex(skey))
	}

	if prev != nil {
		prev.Args = prev.Args.Add(mapKey)
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
	spec      *structSpec
	elemFuncs []parseFunc
}

func (ssp *structSaveParser) parse(pstate *parseState, v reflect.Value) {
	srtKey := newKey(v)
	prev := pstate.popAction()
	curr := &Action{
		Name: "HSET",
		Args: redis.Args{srtKey},
	}
	pstate.pushAction(curr)

	fmt.Printf("ssp: %d\n", len(ssp.spec.fields))
	for i, fldSpec := range ssp.spec.fields {
		fmt.Printf("ssp.parse(%v) \n", fldSpec.typ)
		curr.Args = curr.Args.Add(fldSpec.name)
		ssp.elemFuncs[i](pstate, fldSpec.valueOf(v))
	}

	if prev != nil {
		prev.Args = prev.Args.Add(srtKey)
		pstate.pushAction(prev)
	}
}

func newStructSaveParser(t reflect.Type) parseFunc {
	srtSpec := structSpecForType(t)
	ssp := &structSaveParser{
		spec:      srtSpec,
		elemFuncs: make([]parseFunc, len(srtSpec.fields)),
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
		curr.Args = curr.Args.Add("NULL")
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



type findFunc func(pstate *parseState, v reflect.Value, key string)

func parseFind(v interface{}, key string) {



}

func findParser(t reflect.Type) findFunc {
	return nil
}

type arrayFindParser struct {
	elemFunc findFunc
	handler  ReplyHandler
}

func (afp *arrayFindParser)parse(pstate *parseState, v reflect.Value, key string) {
	action := &Action{
		Name:"LRANGE",
		Args:redis.Args{key, 0, -1},
	}
	pstate.pushAction(action)
}

func newArrayFindParser(t reflect.Type) findFunc {
	afp := &arrayFindParser{findParser(t.Elem()), nil}
	return afp.parse
}
