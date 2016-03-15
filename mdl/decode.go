package mdl

import (
	"fmt"
	"github.com/LindenY/gomdise/trans"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"strconv"
	"sync"
)

type RMNode trans.RMNode

func Decode(node RMNode, dest interface{}) {
	v := reflect.ValueOf(dest)
	decFunc := decoderForType(v.Type())
	decFunc(node, node.Value(), v)
}

type decodeFunc func(node RMNode, data interface{}, v reflect.Value)

var decoderCache struct {
	sync.RWMutex
	m map[reflect.Type]decodeFunc
}

func decoderForType(t reflect.Type) decodeFunc {
	decoderCache.RLock()
	f := decoderCache.m[t]
	decoderCache.RUnlock()
	if f != nil {
		return f
	}

	decoderCache.Lock()
	if decoderCache.m == nil {
		decoderCache.m = make(map[reflect.Type]decodeFunc)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	decoderCache.m[t] = func(node RMNode, data interface{}, v reflect.Value) {
		wg.Wait()
		f(node, data, v)
	}
	decoderCache.Unlock()

	f = newTypeDecoder(t)
	wg.Done()
	decoderCache.Lock()
	decoderCache.m[t] = f
	decoderCache.Unlock()
	return f
}

func newValueForType(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		return reflect.New(t).Elem()
	case reflect.Ptr:
		return reflect.New(t.Elem())
	default:
		return reflect.Zero(t)
	}
}

func newTypeDecoder(t reflect.Type) decodeFunc {
	var decoder decodeFunc
	switch t.Kind() {
	case reflect.Bool:
		decoder = booleanDecoder
	case reflect.String:
		decoder = stringDecoder
	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		decoder = numberDecoder
	case reflect.Map:
		decoder = newMapDecoder(t)
	case reflect.Array, reflect.Slice:
		decoder = newArrayDecoder(t)
	case reflect.Struct:
		decoder = newStructDecoder(t)
	case reflect.Ptr:
		decoder = newPointerDecoder(t)
	default:
		return unsupportedTypeDecoder
	}

	if IfImplementsModel(t) {
		decoder = newModelDecoder(decoder)
	}
	return decoder
}

type arrayDecoder struct {
	elemFunc decodeFunc
}

func (arrDec *arrayDecoder) decode(node RMNode, data interface{}, v reflect.Value) {
	values, err := redis.Values(node.Value(), nil)
	if err != nil {
		panic(err)
	}
	size := len(values)
	v.Set(reflect.MakeSlice(v.Type(), size, size))
	for i := 0; i < size; i++ {
		elemV := newValueForType(v.Type().Elem())
		if i < node.Size() {
			arrDec.elemFunc(node.Child(i), values[i], elemV)
		} else {
			arrDec.elemFunc(node, values[i], elemV)
		}
		v.Index(i).Set(elemV)
	}
}

func newArrayDecoder(t reflect.Type) decodeFunc {
	arrDec := &arrayDecoder{decoderForType(t.Elem())}
	return arrDec.decode
}

type mapDecoder struct {
	elemFunc decodeFunc
}

func (mapDec *mapDecoder) decode(node RMNode, data interface{}, v reflect.Value) {
	values, err := redis.Values(node.Value(), nil)
	if err != nil {
		panic(err)
	}

	size := len(values) / 2
	vals, err := redis.Values(node.Value(), nil)
	if err != nil {
		panic(err)
	}
	v.Set(reflect.MakeMap(v.Type()))

	for i := 0; i < size; i++ {
		mKey, err := redis.String(values[i*2], nil)
		if err != nil {
			panic(err)
		}
		elemV := newValueForType(v.Type().Elem())
		if i < node.Size() {
			mapDec.elemFunc(node.Child(i), vals[i*2+1], elemV)
		} else {
			mapDec.elemFunc(node, vals[i*2+1], elemV)
		}
		v.SetMapIndex(reflect.ValueOf(mKey), elemV)
	}
}

func newMapDecoder(t reflect.Type) decodeFunc {
	mapDec := &mapDecoder{decoderForType(t.Elem())}
	return mapDec.decode
}

type structDecoder struct {
	spec      *StructSpec
	elemFuncs []decodeFunc
}

/*
 * TODO: using key name match instead of matching fields by reply order
 */
func (srtDec *structDecoder) decode(node RMNode, data interface{}, v reflect.Value) {
	values, err := redis.Values(node.Value(), nil)
	if err != nil {
		panic(err)
	}
	size := len(values) / 2
	for i := 0; i < size; i++ {
		fldVal := srtDec.spec.Fields[i].ValueOf(v)
		if i < node.Size() {
			srtDec.elemFuncs[i](node.Child(i), values[i*2+1], fldVal)
		} else {
			srtDec.elemFuncs[i](node, values[i*2+1], fldVal)
		}
	}
}

func newStructDecoder(t reflect.Type) decodeFunc {
	srtSpec := StructSpecForType(t)
	srtDec := &structDecoder{
		spec:      srtSpec,
		elemFuncs: make([]decodeFunc, len(srtSpec.Fields)),
	}

	for i, fld := range srtSpec.Fields {
		srtDec.elemFuncs[i] = decoderForType(fld.Typ)
	}
	return srtDec.decode
}

type pointerDecoder struct {
	elemFunc decodeFunc
}

func (ptrDec *pointerDecoder) decode(node RMNode, data interface{}, v reflect.Value) {
	ptrDec.elemFunc(node, data, v.Elem())
}

func newPointerDecoder(t reflect.Type) decodeFunc {
	ptrDec := &pointerDecoder{decoderForType(t.Elem())}
	return ptrDec.decode
}

type modelDecoder struct {
	elemFunc decodeFunc
}

func (mdlDec *modelDecoder) decode(node RMNode, data interface{}, v reflect.Value) {
	key, err := redis.String(data, nil)
	if err != nil {
		panic(err)
	} else if key != "" && v.CanAddr() {
		ValueSetModelId(v, key)
	}
	mdlDec.elemFunc(node, data, v)
}

func newModelDecoder(dec decodeFunc) decodeFunc {
	mdlDec := &modelDecoder{dec}
	return mdlDec.decode
}

func booleanDecoder(node RMNode, data interface{}, v reflect.Value) {
	val, err := redis.Bool(data, nil)
	if err != nil {
		panic(err)
	}
	v.SetBool(val)
}

func stringDecoder(node RMNode, data interface{}, v reflect.Value) {
	val, err := redis.String(data, nil)
	if err != nil {
		panic(err)
	}
	v.SetString(val)
}

func numberDecoder(node RMNode, data interface{}, v reflect.Value) {
	val, err := redis.String(data, nil)
	if err != nil {
		panic(err)
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil || v.OverflowInt(n) {
			panic(fmt.Sprintf("Unable to convert int: %s \n", val))
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil || v.OverflowUint(n) {
			panic(fmt.Sprintf("Unable to convert uint: %s \n", val))
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			panic(fmt.Sprintf("Unable to convert float: %s \n", val))
		}
		v.SetFloat(n)

	default:
		panic(fmt.Sprintf("Unsupported number convertion for type[%v] with value[%v]", v.Type(), data))
	}
}

func unsupportedTypeDecoder(node RMNode, data interface{}, v reflect.Value) {
	panic(fmt.Sprintf("Unsupported decoding for type[%v] with value[%v]", v.Type(), data))
}
