package gomdies

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
	"github.com/LindenY/gomdise/trans"
	"github.com/LindenY/gomdise/model"
	"fmt"
)

var tcache_find *TemplateCache

func init() {
	tcache_find = newTplCache(newFindTemplateForType)
}


func newFindTemplateForType(t reflect.Type) ActionTemplate {

	fmt.Printf("new tpl for type: %v \n", t)

	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return newVoidFindTemplate(t)
	case reflect.Array, reflect.Slice:
		return newArrayFindTemplate(t)
	case reflect.Map:
		return newMapFindTemplate(t)
	case reflect.Struct:
		return newStructFindTemplate(t)
	case reflect.Ptr:
		return newPointerFindTemplate(t)
	default:
		return nil
	}
}

/*
 *
 */
type arrayFindTemplate struct {
	elemTpl ActionTemplate
}

func (aft *arrayFindTemplate) handle(tran *trans.Transaction, action *trans.Action, reply interface{}) {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		panic(err)
	}
	for _, rpy := range replies {
		newAs := make([]*trans.Action, 0, 1)
		aft.elemTpl.engrave(&newAs, rpy)
		action.AddChildren(newAs...)
		tran.Actions = append(tran.Actions, newAs...)
	}
}

func (aft *arrayFindTemplate) engrave(actions *[]*trans.Action, args ...interface{}){
	key, ok := args[0].(string)
	if !ok {
		_, err := redis.Scan(args, &key)
		if err != nil {
			panic(err)
		}
	}

	action := &trans.Action{
		Name:    "LRANGE",
		Args:    redis.Args{key, 0, -1},
		Handler: aft.handle,
	}
	*actions = append(*actions, action)
}

func newArrayFindTemplate(t reflect.Type) *arrayFindTemplate {
	aft := &arrayFindTemplate{
		elemTpl: tcache_find.GetTemplate(t.Elem()),
	}
	return aft
}

/*
 *
 */
type mapFindTemplate struct {
	elemTpl ActionTemplate
}

func (mft *mapFindTemplate) handle(tran *trans.Transaction, action *trans.Action, reply interface{}) {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		panic(err)
	}

	toggle := false
	for _, rpy := range replies {
		toggle = !toggle
		if toggle {
			continue
		}

		newAs := make([]*trans.Action, 0, 1)
		mft.elemTpl.engrave(&newAs, rpy)
		action.AddChildren(newAs...)
		tran.Actions = append(tran.Actions, newAs...)
	}
}

func (mft *mapFindTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	key, ok := args[0].(string)
	if !ok {
		_, err := redis.Scan(args, &key)
		if err != nil {
			panic(err)
		}
	}
	action := &trans.Action{
		Name:    "HGETALL",
		Args:    redis.Args{key},
		Handler: mft.handle,
	}
	*actions = append(*actions, action)
}

func newMapFindTemplate(t reflect.Type) *mapFindTemplate {
	mft := &mapFindTemplate{
		elemTpl: tcache_find.GetTemplate(t.Elem()),
	}
	return mft
}

/*
 *
 */
type structFindTemplate struct {
	spec    *gomdies.StructSpec
	elemTpl []ActionTemplate
}

func (sft *structFindTemplate) handle(tran *trans.Transaction, action *trans.Action, reply interface{}) {
	replies, err := redis.Values(reply, nil)
	if err != nil {
		panic(err)
	}
	toggle := false
	for i, rpy := range replies {
		toggle = !toggle
		if toggle {
			continue
		}
		newAs := make([]*trans.Action, 0, 1)
		sft.elemTpl[(i-1)/2].engrave(&newAs, rpy)
		action.AddChildren(newAs...)
		tran.Actions = append(tran.Actions, newAs...)
	}
}

func (sft *structFindTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	key, ok := args[0].(string)
	if !ok {
		_, err := redis.Scan(args, &key)
		if err != nil {
			panic(err)
		}
	}

	action := &trans.Action{
		Name:    "HGETALL",
		Args:    redis.Args{key},
		Handler: sft.handle,
	}
	fmt.Printf("appending actions \n")
	*actions = append(*actions, action)
}

func newStructFindTemplate(t reflect.Type) *structFindTemplate {
	srtSpec := gomdies.StructSpecForType(t)
	sft := &structFindTemplate{
		spec:     srtSpec,
		elemTpl: make([]ActionTemplate, len(srtSpec.Fields)),
	}

	for i, fldSpec := range sft.spec.Fields {
		sft.elemTpl[i] = tcache_find.GetTemplate(fldSpec.Typ)
	}
	return sft
}

type pointerFindTemplate struct {
	elemTpl ActionTemplate
}

func (pft *pointerFindTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	pft.elemTpl.engrave(actions, args)
}

func newPointerFindTemplate(t reflect.Type) ActionTemplate {
	pft := &pointerFindTemplate{tcache_find.GetTemplate(t.Elem())}
	return pft
}

type voidFindTemplate struct {
}

func (vft *voidFindTemplate) handle(tran *trans.Transaction, action *trans.Action, reply interface{}) {
}

func (vft *voidFindTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
}

var vft *voidFindTemplate

func newVoidFindTemplate(t reflect.Type) *voidFindTemplate {
	if vft == nil {
		vft = &voidFindTemplate{}
	}
	return vft
}

/*
type UnsupportArgsError struct {
	Msg  string
	Args interface{}
}

func (uaerr UnsupportArgsError) Error() string {
	return fmt.Sprintf("[%v]\t%s", uaerr.Args, uaerr.Msg)
}
*/
