package gomdies

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
	"errors"
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

func parseSave(pstate *parseState, target interface{}) {
	typ := reflect.TypeOf(target)
	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		parseArraySave(pstate, target)
	case reflect.Map:
		parseMapSave(pstate, target)
	case reflect.Struct:
		parseStructSave(pstate, target)
	default:
		parsePrimitiveSave(pstate, target)
	}
}

func parseArraySave(pstate *parseState, target interface{}) {
	tKey := newKey(target)
	prev := pstate.popAction()
	curr := &Action{
		name:"RPUSH",
		args:redis.Args{tKey},
	}
	pstate.pushAction(curr)

	v := reflect.ValueOf(target)
	n := v.Len()
	for i := 0; i < n; i++ {
		parseSave(pstate, v.Index(i).Interface())
	}

	if prev != nil {
		prev.args = prev.args.Add(tKey)
	}
	pstate.pushAction(prev)
}

func parseMapSave(pstate *parseState, target interface{}) {
	tKey := newKey(target)
	prev := pstate.popAction()
	curr := &Action{
		name:"HSET",
		args:redis.Args{tKey},
	}
	pstate.pushAction(curr)

	v := reflect.ValueOf(target)
	mKeys := v.MapKeys()
	for _, mkey := range mKeys {
		curr.args = curr.args.Add(mkey)
		parseSave(pstate, v.MapIndex(mkey).Interface())
	}

	if prev != nil {
		prev.args = prev.args.Add(tKey)
	}
	pstate.pushAction(prev)
}

func parseStructSave(pstate *parseState, target interface{}) {
}

func parsePrimitiveSave(pstate *parseState, target interface{}) {
	if len(pstate.actions) == 0 {
		panic(errors.New("actions is empty"))
	}

	curr := pstate.actions[len(pstate.actions)-1]
	curr.args = curr.args.Add(target)
}


// Save Parse Functions
type parseFunc func(pstate *parseState, v reflect.Value)

func saveParser(t reflect.Type) parseFunc {

}

func newSaveParser(t reflect.Type) {

	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return primitiveSaveParser

	case reflect.Array, reflect.Slice:
		return newArraySaveParser(t)

	case reflect.Map:
		return newMapSaveParser(t)

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
	arrKey := newKey(pstate.target)
	prev := pstate.popAction()
	curr := &Action{
		name: "RPUSH",
		args:redis.Args{arrKey},
	}
	pstate.pushAction(curr)

	n := len(v)
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
	mapKey := newKey(pstate.target)
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

}

func (ssp *structSaveParser) parse(pstate *parseState, v reflect.Value) {

}

func newStructSaveParser(t reflect.Type) parseFunc {
	return nil
}


func unsupportedTypeParser(pstate *parseState, v reflect.Value) {
	panic(&UnsupportedTypeError{v.Type()})
}

type UnsupportedTypeError struct {
	Type reflect.Type
}