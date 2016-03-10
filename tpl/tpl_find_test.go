package tpl

import (
	"testing"
	"reflect"
	"fmt"
	"github.com/LindenY/gomdise/trans"
)

func TestFindTemplateTsB(t *testing.T) {
	tpl := TCFind.GetTemplate(reflect.TypeOf(tsB{}))
	tran := trans.NewTransaction(pool)
	tpl.engrave(&tran.Actions, "gomdies.tsB:0b4063db-81ae-46cc-99e3-b64863caf0ce")
	fmt.Println(tran.Actions)
	tran.Exec()
}