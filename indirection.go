package gomdies

import (
	"reflect"
	"github.com/satori/go.uuid"
)

func newKey(val interface{}) string {
	typ := reflect.TypeOf(val)
	if model, ok := val.(Model); ok {
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
	default:
		key = "_udf"
	}

	uuid := uuid.NewV4()
	return key + uuid.String()
}
