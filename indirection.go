package gomdies

import (
	"reflect"
	"github.com/satori/go.uuid"
	"fmt"
)

var (
	modelType = reflect.TypeOf(new(Model)).Elem()
)

func newKey(val reflect.Value) string {
	if val.Type().Implements(modelType) {
		model := val.Interface().(Model)
		return model.GetModelId()
	}

	uuid := uuid.NewV4()
	return fmt.Sprintf("%v:%s", val.Type(), uuid.String())
}
