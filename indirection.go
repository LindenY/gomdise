package gomdies

import (
	"reflect"
	"github.com/satori/go.uuid"
)

var (
	modelType = reflect.TypeOf(new(Model)).Elem()
)

func newKey(val reflect.Value) string {
	typ := reflect.TypeOf(val)
	if val.Type().Implements(modelType) {
		model := val.Interface().(Model)
		return model.GetModelId()
	}

	var key string
	switch typ.Kind() {
	case reflect.Array, reflect.Slice:
		key = "_arr:"
	case reflect.Map:
		key = "_map:"
	case reflect.Struct:
		key = "_srt:"
	case reflect.Ptr:
		key = newKey(val.Elem())
	default:
		key = "_udf"
	}

	uuid := uuid.NewV4()
	return key + uuid.String()
}
