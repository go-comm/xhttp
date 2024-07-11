package xhttp

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type BindTextUnmarshaler interface {
	UnmarshalText(text []byte) error
}

func BindQuery(r http.Request, v interface{}) error {
	var query map[string][]string
	if r.URL != nil {
		query = r.URL.Query()
	}
	if err := BindData(v, query, ""); err != nil {
		return NewHttpError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func BindForm(r http.Request, v interface{}) error {
	if err := BindData(v, r.Form, ""); err != nil {
		return NewHttpError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func BindHeaders(r http.Request, v interface{}) error {
	if err := BindData(v, r.Header, ""); err != nil {
		return NewHttpError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func BindData(v interface{}, values map[string][]string, tag string) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	m := OpenReflectMapper(rv.Type(), tag)
	for k, v := range values {
		if len(v) > 0 {
			fv, ok := m.FieldByName(rv, k)
			if !ok && len(tag) > 0 {
				fv, ok = m.FieldByTag(rv, k)
			}
			if ok {
				if err := setFieldValue(fv, v[0]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setFieldValue(rv reflect.Value, v string) error {
	switch rv.Kind() {
	case reflect.Bool:
		v = strings.ToUpper(v)
		if v == "TRUE" || v == "T" || v == "1" {
			rv.SetBool(true)
		} else if v == "FALSE" || v == "F" || v == "0" {
			rv.SetBool(false)
		} else {
			return fmt.Errorf("can not read field %v to %v", v, rv.Kind())
		}
	case reflect.String:
		rv.SetString(v)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		n, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return err
		}
		rv.SetInt(n)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		n, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return err
		}
		rv.SetUint(n)
	case reflect.Float32:
		n, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return err
		}
		rv.SetFloat(n)
	case reflect.Float64:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(n)
	case reflect.Struct, reflect.Ptr:
		if um, ok := rv.Interface().(BindTextUnmarshaler); ok {
			return um.UnmarshalText(StrToBytes(v))
		}
	default:
		return fmt.Errorf("can not read field %v to %v", v, rv.Kind())
	}
	return nil
}
