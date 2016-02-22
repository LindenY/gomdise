package gomdies

import (
	"testing"
	"reflect"
	"fmt"
)


type tsA struct {
	tsB
	IntVal int
	StrVal string
	SrtVal tsB
	PtrVal *tsB
	MapVal map[string]string
}

type tsB struct {
	BytVal byte
	LstVal []int
}


func TestScan(t *testing.T) {

	tsb0 := tsB{
		BytVal:0x00,
		LstVal:make([]int, 3),
	}

	tsb1 := tsB{
		BytVal:0x01,
		LstVal:make([]int, 2),
	}

	tsa0 := tsA{
		IntVal: 0,
		StrVal: "tsa0",
		SrtVal:tsb0,
		PtrVal:&tsb1,
		MapVal:make(map[string]string),
	}

	sSpec := structSpecForType(reflect.TypeOf(tsa0))

	fmt.Printf("Struct:%v\t%d fields \n", reflect.TypeOf(tsa0), len(sSpec.fields))
	for i, fSpec := range sSpec.fields {
		fmt.Printf("[%d]\t%s:\t%v\n", i, fSpec.name, fSpec)
	}
}
