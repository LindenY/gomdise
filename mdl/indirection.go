package mdl

import (
	"fmt"
	"github.com/satori/go.uuid"
	"reflect"
)

var (
	modelType = reflect.TypeOf(new(Model)).Elem()
)

func NewKey(val reflect.Value) string {
	if IfImplementsModel(val.Type()) && val.CanAddr() {
		return ValueGetModelId(val)
	}

	switch val.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return ""
	}

	uuid := uuid.NewV4()
	return fmt.Sprintf("%v:%s", val.Type(), uuid.String())
}
