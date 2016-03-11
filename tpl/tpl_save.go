package tpl

import (
	"reflect"
	"github.com/garyburd/redigo/redis"
	"github.com/LindenY/gomdise/mdl"
	"github.com/LindenY/gomdise/trans"
	"fmt"
)

var TCSave *TemplateCache

func init() {
	TCSave = newTplCache(newSaveTemplateForType)
	_prtst = &primitiveSaveTemplate{}
}

func newSaveTemplateForType(t reflect.Type) ActionTemplate {
	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return _prtst
	case reflect.Array, reflect.Slice:
		return newArraySaveTemplate(t)
	case reflect.Map:
		return newMapSaveTemplate(t)
	case reflect.Struct:
		return newStructSaveTemplate(t)
	case reflect.Ptr:
		return newPointerSaveTemplate(t)
	default:
		return newUnsupportedTypeTemplate(t, "Save")
	}
}

/*
 *
 */
type arraySaveTemplate struct {
	elemTpl ActionTemplate
}

func (ast *arraySaveTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	action := &trans.Action {
		Name:"RPUSH",
		Args:redis.Args{args[0]},
	}
	*actions = append(*actions, action)

	v := args[1].(reflect.Value)
	n := v.Len()
	for i := 0; i < n; i ++ {
		eKey := mdl.NewKey(v.Index(i))
		if eKey != "" {
			action.Args = action.Args.Add(eKey)
		}
		ast.elemTpl.Engrave(actions, eKey, v.Index(i))
	}
}

func newArraySaveTemplate(t reflect.Type) *arraySaveTemplate {
	ast := &arraySaveTemplate{TCSave.GetTemplate(t.Elem())}
	return ast
}

/*
 *
 */
type mapSaveTemplate struct {
	elemTpl ActionTemplate
}

func (mst *mapSaveTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	action := &trans.Action {
		Name:"HMSET",
		Args:redis.Args{args[0]},
	}
	*actions = append(*actions, action)

	v := args[1].(reflect.Value)
	mKey := v.MapKeys()
	for _, mKey := range mKey {
		action.Args = action.Args.Add(mKey)
		mVal := mdl.NewKey(v.MapIndex(mKey))
		if mVal != "" {
			action.Args = action.Args.Add(mVal)
		}
		mst.elemTpl.Engrave(actions, mVal, v.MapIndex(mKey))
	}
}

func newMapSaveTemplate(t reflect.Type) *mapSaveTemplate {
	mst := &mapSaveTemplate{TCSave.GetTemplate(t.Elem())}
	return mst
}

/*
 *
 */
type structSaveTemplate struct {
	spec *mdl.StructSpec
	elemTpls []ActionTemplate
}

func (sst *structSaveTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	action := &trans.Action {
		Name:"HMSET",
		Args:redis.Args{args[0]},
	}
	*actions = append(*actions, action)

	v := args[1].(reflect.Value)
	for i, fld := range sst.spec.Fields {
		action.Args = action.Args.Add(fld.Name)
		fVal := fld.ValueOf(v)
		fKey := mdl.NewKey(fVal)
		if fKey != "" {
			action.Args = action.Args.Add(fKey)
		}
		sst.elemTpls[i].Engrave(actions, fKey, fVal)
	}
}

func newStructSaveTemplate(t reflect.Type) *structSaveTemplate {
	spec := mdl.StructSpecForType(t)
	sst := &structSaveTemplate{
		spec: spec,
		elemTpls:make([]ActionTemplate, len(spec.Fields)),
	}
	for i, fld := range sst.spec.Fields {
		sst.elemTpls[i] = TCSave.GetTemplate(fld.Typ)
	}
	return sst
}

/*
 *
 */
type pointerSaveTemplate struct {
	elemFunc ActionTemplate
}

func (pst *pointerSaveTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	fmt.Printf("%v \n", args)
	v := args[1].(reflect.Value)

	pst.elemFunc.Engrave(actions, args[0], v.Elem())
}

func newPointerSaveTemplate(t reflect.Type) *pointerSaveTemplate {
	pst := &pointerSaveTemplate{TCSave.GetTemplate(t.Elem())}
	return pst
}

/*
 *
 */
type primitiveSaveTemplate struct {}
var _prtst *primitiveSaveTemplate

func (pst *primitiveSaveTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	action := (*actions)[len(*actions) - 1]
	action.Args = action.Args.Add(args[1].(reflect.Value).Interface())
}
