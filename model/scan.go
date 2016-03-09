package gomdies

import (
	"reflect"
	"sync"
)

type FieldSpec struct {
	Name  string
	Index []int
	Tag   bool
	Typ   reflect.Type
}

func (fldSpec *FieldSpec) valueOf(v reflect.Value) reflect.Value {
	retVal := v
	for _, fldIdx := range fldSpec.Index {
		retVal = retVal.Field(fldIdx)
	}
	return retVal
}

type StructSpec struct {
	Fields []*FieldSpec
}

func compileStructSpec(t reflect.Type) *StructSpec {

	current := []FieldSpec{}
	next := []FieldSpec{{Typ: t}}

	visited := map[reflect.Type]bool{}

	fieldSpecs := make([]*FieldSpec, 0)
	for len(next) > 0 {
		current, next = next, current[:0]

		for _, fs := range current {
			if visited[fs.Typ] {
				continue
			}
			visited[fs.Typ] = true

			// Scan fs.type for fields to include
			for i := 0; i < fs.Typ.NumField(); i++ {
				sfs := fs.Typ.Field(i)

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

				index := make([]int, len(fs.Index)+1)
				copy(index, fs.Index)
				index[len(fs.Index)] = i

				ft := sfs.Type
				for ft.Name() == "" && ft.Kind() == reflect.Ptr { // follow the pointer
					ft = ft.Elem()
				}

				if name != "" || !sfs.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sfs.Name
					}
					fieldSpecs = append(fieldSpecs, &FieldSpec{
						Name:  name,
						Tag:   tagged,
						Index: index,
						Typ:   ft,
					})
					continue
				}

				next = append(next, FieldSpec{
					Name:  ft.Name(),
					Index: index,
					Typ:   ft,
				})
			}
		}
	}

	return &StructSpec{fieldSpecs}
}

var structSpecCache struct {
	sync.RWMutex
	m map[reflect.Type]*StructSpec
}

func StructSpecForType(t reflect.Type) *StructSpec {

	structSpecCache.RLock()
	ss := structSpecCache.m[t]
	structSpecCache.RUnlock()

	if ss != nil {
		return ss
	}

	ss = compileStructSpec(t)
	structSpecCache.Lock()
	if structSpecCache.m == nil {
		structSpecCache.m = map[reflect.Type]*StructSpec{}
	}
	structSpecCache.m[t] = ss
	structSpecCache.Unlock()

	return ss
}
