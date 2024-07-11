package xhttp

import (
	"reflect"
	"strings"
	"sync"
)

var (
	glabelReflectMappers sync.Map
	defaultDeep          = 999
	zeroValue            = reflect.Value{}
)

type keyOfReflectMapper struct {
	rtype reflect.Type
	tag   string
}

func OpenReflectMapper(t reflect.Type, tagName string) *ReflectMapper {
	k := keyOfReflectMapper{rtype: t, tag: tagName}
	var m *ReflectMapper
	v, ok := glabelReflectMappers.Load(k)
	if ok {
		m, ok = v.(*ReflectMapper)
	}
	if ok {
		return m
	}
	m = newReflectMapper(t, tagName)
	glabelReflectMappers.Store(k, m)
	return m
}

type FieldInfo struct {
	Index       []int
	StructField reflect.StructField
}

type ReflectMapper struct {
	fields []*FieldInfo
	named  map[string]*FieldInfo
	tagged map[string]*FieldInfo
}

func newReflectMapper(t reflect.Type, tagName string) *ReflectMapper {
	fields := ListFields(t)
	named := map[string]*FieldInfo{}
	tagged := map[string]*FieldInfo{}
	for _, f := range fields {
		named[f.StructField.Name] = f
	}
	if len(tagName) > 0 {
		for _, f := range fields {
			tv := f.StructField.Tag.Get(tagName)
			if p := strings.Index(tv, ","); p >= 0 {
				tv = tv[:p]
			}
			tagged[tv] = f
		}
	}
	return &ReflectMapper{fields: fields, named: named, tagged: tagged}
}

func (m *ReflectMapper) Fields() []*FieldInfo {
	return m.fields
}

func (m *ReflectMapper) FieldInfoByName(name string) (*FieldInfo, bool) {
	fi, ok := m.named[name]
	return fi, ok
}

func (m *ReflectMapper) FieldByName(v reflect.Value, name string) (reflect.Value, bool) {
	fi, ok := m.named[name]
	if !ok {
		return zeroValue, false
	}
	return FieldByIndex(v, fi.Index), true
}

func (m *ReflectMapper) FieldInfoByTag(tagValue string) (*FieldInfo, bool) {
	fi, ok := m.tagged[tagValue]
	return fi, ok
}

func (m *ReflectMapper) FieldByTag(v reflect.Value, tagValue string) (reflect.Value, bool) {
	fi, ok := m.tagged[tagValue]
	if !ok {
		return zeroValue, false
	}
	return FieldByIndex(v, fi.Index), true
}

func rangeFields(t reflect.Type, index []int, deep int, fn func(fi *FieldInfo) (continued bool)) bool {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fi := &FieldInfo{
			Index:       append(append(([]int)(nil), index...), f.Index...),
			StructField: f,
		}
		if !fn(fi) {
			return false
		}
		if deep > 0 && f.Anonymous && f.Type.Kind() == reflect.Struct {
			if !rangeFields(f.Type, fi.Index, deep-1, fn) {
				return false
			}
		}
	}
	return true
}

func ListFields(t reflect.Type) []*FieldInfo {
	var ls []*FieldInfo
	rangeFields(t, nil, defaultDeep, func(fi *FieldInfo) (continued bool) {
		ls = append(ls, fi)
		return true
	})
	return ls
}

func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func FieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		v = reflect.Indirect(v).Field(i)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			v.Set(reflect.New(Deref(v.Type())))
		}
		if v.Kind() == reflect.Map && v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
	}
	return v
}
