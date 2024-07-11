package xhttp

import (
	"reflect"
	"testing"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Student struct {
	Person
	Class string `json:"class,omitempty"`
}

func TestFieldByNameOfMapper(t *testing.T) {

	st := &Student{
		Person: Person{
			Name: "john",
			Age:  12,
		},
		Class: "Class One",
	}
	v := reflect.Indirect(reflect.ValueOf(st))
	m := OpenReflectMapper(v.Type(), "json")
	want := reflect.Indirect(reflect.ValueOf(&st.Age))

	if got, _ := m.FieldByName(v, "Age"); got != want {
		t.Fatal(want, got)
	}
	if got, _ := m.FieldByTag(v, "age"); got != want {
		t.Fatal(want, got)
	}

}
