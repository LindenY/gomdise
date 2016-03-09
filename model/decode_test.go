package gomdies

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDecodeTsB(t *testing.T) {

	tpl := findTemplateForType(reflect.TypeOf(tsB{}))
	findAction := tpl.engrave("gomdies.tsB:0b4063db-81ae-46cc-99e3-b64863caf0ce")

	tran := NewTransaction(pool)
	tran.Actions = append(tran.Actions, findAction)
	printActions(tran.Actions)
	tran.Exec()

	tsB := tsB{}
	Decode(findAction, &tsB)

	fmt.Printf("%v \n", tsB)
}
