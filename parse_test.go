package gomdies

import (
	"testing"
	"fmt"
)

type tsA struct {
	m map[string][]int
	tsB
}

type tsB struct {
	m map[string]*tsC
}

type tsC struct {
	str string
	ls []int
}

func TestParse(t *testing.T) {
	l0 := make([]int, 3)
	l0[0] = 3
	l0[1] = 4
	l0[2] = 5

	l1 := make([]int, 2)
	l1[0] = 6
	l1[1] = 7

	tsC0 := &tsC{"tsC0", l0}
	tsC1 := &tsC{"tsC1", l1}
	m0 := make(map[string]*tsC)
	m0["tsC0"] = tsC0
	m0["tsC1"] = tsC1

	tsB0 := tsB{m0}

	m1 := make(map[string][]int)
	m1["l0"] = l0
	m1["l1"] = l1

	tsA0 := tsA{
		m: m1,
		tsB: tsB0,
	}


	actions := parseSave(tsA0)

	fmt.Printf("%v\n", saveParserCache.m)
	fmt.Println()
	printActions(actions)

}


func printActions(actions  []*Action) {
	fmt.Printf("Num of actions: %d\n", len(actions))
	for i, a := range actions {
		fmt.Printf("\t[%d]:\t%v\n", i, a)
	}
}
