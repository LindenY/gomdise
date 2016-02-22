package gomdies

import (
	"testing"
	"fmt"
)

type ptsA struct {
	ptsB
	M map[string][]int
}

type ptsB struct {
	M map[string]*ptsC
}

type ptsC struct {
	Str string
	Ls []int
}

func TestParse(t *testing.T) {
	l0 := make([]int, 3)
	l0[0] = 3
	l0[1] = 4
	l0[2] = 5

	l1 := make([]int, 2)
	l1[0] = 6
	l1[1] = 7

	tsC0 := &ptsC{"tsC0", l0}
	tsC1 := &ptsC{"tsC1", l1}
	m0 := make(map[string]*ptsC)
	m0["tsC0"] = tsC0
	m0["tsC1"] = tsC1

	m1 := make(map[string][]int)
	m1["l0"] = l0
	m1["l1"] = l1

	tsA0 := ptsA{
		M: m1,
		ptsB: ptsB{m0},
	}


	actions, err := parseSave(tsA0)
	if err != nil {
		panic(err)
	}

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
