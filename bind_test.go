package xhttp

import (
	"net/url"
	"testing"
)

func TestBindData(t *testing.T) {

	var values url.Values
	values, _ = url.ParseQuery("user=join&age=21&Class=zh")
	var user struct {
		User  string `json:"user,omitempty"`
		Age   int    `json:"age,omitempty"`
		Class string
	}
	if err := BindData(&user, values, "json"); err != nil {
		t.Fatal(err)
	}
	pass := user.User == "join" && user.Age == 21 && user.Class == "zh"
	if !pass {
		t.Fatal("wrong user", user)
	}
}
