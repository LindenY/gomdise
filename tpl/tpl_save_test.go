package tpl

import (
	"testing"
	"reflect"
	"github.com/LindenY/gomdise/trans"
	"github.com/LindenY/gomdise/mdl"
	"fmt"
)

func TestSaveTemplateTsB(t *testing.T) {
	tsB := MakeTsB()
	tpl := TCSave.GetTemplate(reflect.TypeOf(tsB))
	trans := trans.NewTransaction(pool)
	tpl.Engrave(&trans.Actions, mdl.NewKey(reflect.ValueOf(tsB)), reflect.ValueOf(tsB))

	printActions(trans.Actions)
	trans.Exec()
}

func printActions(actions []*trans.Action) {
	fmt.Printf("Num of actions: %d\n", len(actions))
	for i, a := range actions {
		fmt.Printf("\t[%d]:\t%v\n", i, a)
	}
}