package tpl

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
	"github.com/LindenY/gomdise/trans"
	"github.com/LindenY/gomdise/mdl"
	"github.com/LindenY/gomdise/lscp"
)

var (
	FindTemplate ActionTemplate
)

func init() {

}

type findTemplate struct {

}

func(ft *findTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {

}

func(ft *findTemplate) handle(tran *trans.Transaction, action *trans.Action, reply interface{}) {

}