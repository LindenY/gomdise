package gomdies

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
)

func findTemplateForType(t reflect.Type) {

}

func newFindTemplateForType(t reflect.Type) ActionTemplate {
	return nil
}


type arrayFindTemplate struct {
	ActionTemplate
	elemEgr ArgsEngraver
}

func newArrayFindTemplate(t reflect.Type) ActionTemplate {
	aft := arrayFindTemplate{
		ActionTemplate.Name: "LRANGE",
	}

	aft.ActionTemplate.Handler = func(tran *Transaction, reply interface{}) error {
		replies, err := redis.Values(reply, nil)
		if err != nil {
			return err
		}
		for _, rpy := range replies {
			err = aft.elemEgr(tran, rpy)
			if err != nil {
				return err
			}
		}
		return nil
	}

	aft.ActionTemplate.Engraver = func(tran *Transaction, args interface{}) error {
		action := &Action{aft}

	}
}
