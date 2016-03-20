package tpl

import (
	"reflect"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/LindenY/gomdise/mdl"
	"github.com/LindenY/gomdise/trans"
)

var (
	TCSave *TemplateCache
 	_prtst *primitiveSaveTemplate
	_infst *interfaceSaveTemplate
)

func init() {
	TCSave = newTplCache(newSaveTemplateForType)
	_prtst = &primitiveSaveTemplate{}
	_infst = &interfaceSaveTemplate{}
}

/*
 * TODO: save template can not work with interface type
 *  For innstance, when save []interface{} type, then
 *  the underlying data type is mixed with primitive
 *  types and others, the template with not work. Since
 *  the primitive template will add its value to last
 *  action, but the last action might not be the action
 *  of the []interface type itself, but of others data type.
 */
func newSaveTemplateForType(t reflect.Type) ActionTemplate {
	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return _prtst
	case reflect.Interface:
		return _infst
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

func (ast *arraySaveTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	action := &trans.Action{
		Name: "RPUSH",
		Args: redis.Args{args[0]},
	}
	*actions = append(*actions, action)

	v := args[1].(reflect.Value)
	n := v.Len()
	for i := 0; i < n; i++ {
		eKey := mdl.NewKey(v.Index(i))
		if eKey != "" {
			action.Args = action.Args.Add(eKey)
		}
		ast.elemTpl.Engrave(actions, eKey, v.Index(i))
	}

	action.Handler = func(tran *trans.Transaction, action *trans.Action, reply interface{}) {
		numbericReplyValidator(action, n, reply)
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

func (mst *mapSaveTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	action := &trans.Action{
		Name: "HMSET",
		Args: redis.Args{args[0]},
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

	action.Handler = func(tran *trans.Transaction, action *trans.Action, reply interface{}) {
		stringReplyValidator(action, "OK", reply)
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
	spec     *mdl.StructSpec
	elemTpls []ActionTemplate
}

func (sst *structSaveTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	action := &trans.Action{
		Name: "HMSET",
		Args: redis.Args{args[0]},
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

	action.Handler = func(tran *trans.Transaction, action *trans.Action, reply interface{}) {
		stringReplyValidator(action, "OK", reply)
	}
}

func newStructSaveTemplate(t reflect.Type) *structSaveTemplate {
	spec := mdl.StructSpecForType(t)
	sst := &structSaveTemplate{
		spec:     spec,
		elemTpls: make([]ActionTemplate, len(spec.Fields)),
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
	elemTpl ActionTemplate
}

func (pst *pointerSaveTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	v := args[1].(reflect.Value)
	pst.elemTpl.Engrave(actions, args[0], v.Elem())
}

func newPointerSaveTemplate(t reflect.Type) *pointerSaveTemplate {
	pst := &pointerSaveTemplate{TCSave.GetTemplate(t.Elem())}
	return pst
}

/*
 *
 */
type interfaceSaveTemplate struct {}

func (ist *interfaceSaveTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	v := args[1].(reflect.Value)
	v = v.Elem()
	tpl := TCSave.GetTemplate(v.Type())

	fmt.Printf("interface.Engrave: [args: %v] \n", args)
	fmt.Printf("interface.Engrave: [%v] %v \n", reflect.ValueOf(tpl).Elem().Type(), v)
	tpl.Engrave(actions, args[0], v)
}

/*
 *
 */
type primitiveSaveTemplate struct{}

func (pst *primitiveSaveTemplate) Engrave(actions *[]*trans.Action, args ...interface{}) {
	action := (*actions)[len(*actions)-1]
	fmt.Printf("primitiveTemplate.engrave: [args %d: %v] \n", len(args), args)
	fmt.Printf("primitiveTemplate.engrave: %v \n", args[1].(reflect.Value).Interface())
	fmt.Printf("action.args = %d : ", len(action.Args))
	action.Args = action.Args.Add(args[1].(reflect.Value).Interface())
	fmt.Printf("%d \n", len(action.Args))
}

func numbericReplyValidator(action *trans.Action, ept int, reply interface{}) {
	if rpy, _ := redis.Int(reply, nil); rpy != ept {
		panic(errors.New(fmt.Sprintf("Numberic Reply Error: action[%v] is expecting %d but receviced %v", action, ept, reply)))
	}
}

func stringReplyValidator(action *trans.Action, ept string, reply interface{}) {
	if rpy, _ := reply.(string); rpy != ept {
		panic(errors.New(fmt.Sprintf("String Reply Error: action[%v] is expecting %s but receviced %v", action, ept, reply)))
	}
}