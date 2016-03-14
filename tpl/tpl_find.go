package tpl

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
	"github.com/LindenY/gomdies/trans"
	"github.com/LindenY/gomdies/mdl"
)

var TCFind *TemplateCache

func init() {
	TCFind = newTplCache(newFindTemplateForType)
}


func newFindTemplateForType(t reflect.Type) ActionTemplate {
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
		return newUnsupportedTypeTemplate(t, "Find")
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
		aft.elemTpl.Engrave(&newAs, rpy)
		action.AddChildren(newAs...)
		tran.Actions = append(tran.Actions, newAs...)
	}
}

func (aft *arrayFindTemplate) Engrave(actions *[]*trans.Action, args ...interface{}){
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
		elemTpl: TCFind.GetTemplate(t.Elem()),
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
		mft.elemTpl.Engrave(&newAs, rpy)
		action.AddChildren(newAs...)
		tran.Actions = append(tran.Actions, newAs...)
	}
}

func (mft *mapFindTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
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
		elemTpl: TCFind.GetTemplate(t.Elem()),
	}
	return mft
}

/*
 *
 */
type structFindTemplate struct {
	spec    *mdl.StructSpec
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
		sft.elemTpl[(i-1)/2].Engrave(&newAs, rpy)
		action.AddChildren(newAs...)
		tran.Actions = append(tran.Actions, newAs...)
	}
}

func (sft *structFindTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
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
	*actions = append(*actions, action)
}

func newStructFindTemplate(t reflect.Type) *structFindTemplate {
	srtSpec := mdl.StructSpecForType(t)
	sft := &structFindTemplate{
		spec:     srtSpec,
		elemTpl: make([]ActionTemplate, len(srtSpec.Fields)),
	}

	for i, fldSpec := range sft.spec.Fields {
		sft.elemTpl[i] = TCFind.GetTemplate(fldSpec.Typ)
	}
	return sft
}

type pointerFindTemplate struct {
	elemTpl ActionTemplate
}

func (pft *pointerFindTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	pft.elemTpl.Engrave(actions, args...)
}

func newPointerFindTemplate(t reflect.Type) ActionTemplate {
	pft := &pointerFindTemplate{TCFind.GetTemplate(t.Elem())}
	return pft
}

type voidFindTemplate struct {
}

func (vft *voidFindTemplate) handle(tran *trans.Transaction, action *trans.Action, reply interface{}) {
}

func (vft *voidFindTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
}

var _vft *voidFindTemplate

func newVoidFindTemplate(t reflect.Type) *voidFindTemplate {
	if _vft == nil {
		_vft = &voidFindTemplate{}
	}
	return _vft
}
