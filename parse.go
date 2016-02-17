package gomdies

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
)

type parseState struct {
	actions []*Action
	target *interface{}
}


func (pstate *parseState) pushAction(action *Action) {
	if action == nil {
		return
	}

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


func parseSave(pstate *parseState, target *interface{}) {
	typ := reflect.TypeOf(target)
	switch typ.Kind() {
	case reflect.Slice || reflect.Array:
		parseArraySave(pstate, target)
	}
}

func parseArraySave(pstate *parseState, target *interface{}) {
	tKey := newKey(target)
	prev := pstate.popAction()
	curr := &Action{
		name:"RPUSH",
		args:redis.Args{tKey},
	}
	pstate.pushAction(curr)

	v := reflect.ValueOf(target)
	n := v.Len()
	for i := 1; i < n; i++ {
		parseSave(pstate, target)
	}

	if prev != nil {
		prev.args.Add(tKey)
	}
	pstate.pushAction(prev)
}

func parseMapSave(pstate *parseState, target *interface{}) {
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
		parseSave(pstate, v.MapIndex(mKey))
	}

	if prev != nil {
		prev.args.Add(tKey)
	}
	pstate.pushAction(prev)
}

func parseStructSave(pstate *parseState, target *interface{}) {

}

func parsePrimitiveSave(pstate *parseState, target *interface{}) {
	curr := pstate.actions[len(pstate.actions)-1]
	curr.args.Add(target)
}