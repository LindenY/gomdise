package mdl

import (
	"reflect"
)

type Model interface {
	GetModelId() string
	SetModelId(id string)
}

var _modelType reflect.Type

func init() {
	_modelType = reflect.TypeOf((*Model)(nil)).Elem()
}

func IfImplementsModel(t reflect.Type) bool {
	if t.Implements(_modelType) {
		return true
	}
	if t.Kind() != reflect.Ptr {
		if reflect.PtrTo(t).Implements(_modelType) {
			return true
		}
	}
	return false
}

func ValueGetModelId(v reflect.Value) string {
	var va reflect.Value
	if v.Kind() != reflect.Ptr {
		va = v.Addr()
	} else {
		va = v
	}
	if va.IsNil() {
		return ""
	}
	mdl := va.Interface().(Model)
	return mdl.GetModelId()
}

func ValueSetModelId(v reflect.Value, key string) {
	var va reflect.Value
	if v.Kind() != reflect.Ptr {
		va = v.Addr()
	} else {
		va = v
	}
	if va.IsNil() {
		return
	}
	method := va.MethodByName("SetModelId")
	method.Call([]reflect.Value{reflect.ValueOf(key)})
}
