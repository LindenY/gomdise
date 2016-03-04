package gomdies

import (
	_ "github.com/garyburd/redigo/redis"
	_ "os"
	"reflect"
	"testing"
	_ "time"
)

func TestFindTemplateTsB(t *testing.T) {

	tpl := findTemplateForType(reflect.TypeOf(tsB{}))

	tran := NewTransaction(pool)
	tran.Actions = append(tran.Actions, tpl.engrave("gomdies.tsB:0b4063db-81ae-46cc-99e3-b64863caf0ce"))
	printActions(tran.Actions)
	tran.Exec()
}
