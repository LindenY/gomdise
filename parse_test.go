package gomdies

import (
	"testing"
	_"github.com/garyburd/redigo/redis"
	"fmt"
	"reflect"
)

func TestParse(t *testing.T) {

	pstate := &parseState{
		actions:make([]*Action, 0),
	}

	l0 := make([]int, 3)
	l0[0] = 3
	l0[1] = 4
	l0[2] = 5

	l1 := make([]int, 2)
	l1[0] = 6
	l1[1] = 7

	m := make(map[string][]int)
	m["l0"] = l0
	m["l1"] = l1


	pf := saveParser(reflect.TypeOf(m))
	pf(pstate, reflect.ValueOf(m))

	fmt.Printf("%v\n", parserCache.m)
	fmt.Println()
	fmt.Printf("Num of actions: %d\n", len(pstate.actions))
	for i, a := range pstate.actions {
		fmt.Printf("\t[%d]:\t%v \n", i, a)
	}

}


func printActions(actions  []*Action) {

	fmt.Sprintf("Num of actions: %d", len(actions))
	for i, a := range actions {
		fmt.Sprintf("\t[%d]:\t%v", i, a)
	}
}
