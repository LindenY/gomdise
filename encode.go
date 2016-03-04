package gomdies

import (
	_ "reflect"
	_ "sync"
)

/*
type encoderFunc func(script *Script, v reflect.Value)


var encoderCache struct {
	sync.RWMutex
	m map[reflect.Type]encoderFunc
}

func typeEncoder(typ reflect.Type) encoderFunc {

	encoderCache.RLock()
	ef := encoderCache.m[typ]
	encoderCache.RUnlock()
	if (ef != nil) {
		return ef
	}

	ef = newTypeEncoder(typ)
	encoderCache.Lock()
	encoderCache.m[typ] = ef
	encoderCache.Unlock()
	return ef
}

func newTypeEncoder(typ reflect.Type) encoderFunc {

	switch typ.Kind() {
	case reflect.Struct:

	default:
		return unsupportedTypeEncoder
	}
}

func unsupportedTypeEncoder(script Script, v reflect.Value) {
	script.e = error("unsupported type")
}


type structEncoder struct {
	spec structSpec
	efs []encoderFunc
}

func (se *structEncoder) flattern(v reflect.Value) Args {

	args := make([]interface{}, len(se.spec.fmap))

	for fname, fspec := range se.spec.fmap {
		fvalue := v.FieldByIndex(fspec.index)
		args = append(args, fname, fvalue.Interface())
	}

}

func (se *structEncoder) encode(script *Script, v reflect.Value) {

}


func newStructEncoder(typ reflect.Type) encoderFunc {
	ss := structSpecForType(typ)

}
*/
