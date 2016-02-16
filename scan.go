package gomdies

import (
	"reflect"
	"sync"
)


type fieldSpec struct {
	name string
	index []int
	tag bool
	typ reflect.Type
}

type structSpec struct {
	fmap map[string]*fieldSpec
}

func compileStructSpec(t reflect.Type) *structSpec {

	current := []fieldSpec{}
	next := []fieldSpec{ {typ:t} }

	visited := map[reflect.Type]bool{}

	fieldSpecs := make(map[string]fieldSpec)

	for (len(current) > 0) {
		current, next = next, current[:0]

		for _, fs := range current {
			if visited[fs.typ] {
				continue
			}
			visited[fs.typ] = true

			// Scan fs.type for fields to include
			for i := 0; i < fs.typ.NumField(); i++ {
				sfs := fs.typ.Field(i);
				if sfs.PkgPath != "" && !sfs.Anonymous { // unexported
					continue
				}

				tag := sfs.Tag.Get("redis")
				if tag == "-" {
					continue
				}
				name, _ := parseTag(tag)
				if !isValidTag(name) {
					name = ""
				}

				index := make([]int, len(fs.index) + 1)
				copy(index, fs.index)
				index[len(fs.index)] = i

				ft := sfs.Type
				for ft.Name() == "" && ft.Kind() == reflect.Ptr { // follow the pointer
					ft = ft.Elem()
				}

				if name != "" || !sfs.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sfs.Name()
					}
					fieldSpecs[name] = *fieldSpec{
						name: name,
						tag: tagged,
						index: index,
						typ: ft,
					}
					continue
				}

				next = append(next, fieldSpec{
					name: ft.Name(),
					index: index,
					typ: ft,
				})
			}
		}
	}

	return &structSpec{fieldSpecs}
}

func compileFieldSpec(structf *reflect.StructField) *fieldSpec {
	tag := structf.Tag.Get("redis")
	if tag == "-" {
		return nil
	}

	name, _ := parseTag(tag)
	if !isValidTag(name) {
		name = ""
	}

	typ := structf.Type
	for (typ.Name() == "" && typ.Kind() == reflect.Ptr) {
		typ = typ.Elem()
	}

	tagged := name != ""
	if name == "" {
		name = typ.Name()
	}

	return &fieldSpec{
		name: name,
		tag: tagged,
		typ: typ,
	}
}

var structSpecCache struct {
	sync.RWMutex
	m map[reflect.Type]*structSpec
}

func structSpecForType(t reflect.Type) structSpec {

	structSpecCache.RLock()
	ss := structSpecCache.m[t]
	structSpecCache.RUnlock()

	if (ss != nil) {
		return ss
	}

	ss = compileStructSpec(t)
	structSpecCache.Lock()
	if structSpecCache.m == nil {
		structSpecCache.m = map[reflect.Type]*structSpec{}
	}
	structSpecCache.m[t] = ss
	structSpecCache.Unlock()

	return ss
}