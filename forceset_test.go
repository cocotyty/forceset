package forceset

import (
	"reflect"
	"testing"
)

type x int

func TestForceSetAlias(t *testing.T) {
	var i int = 0
	v := reflect.ValueOf(&i)

	err := ForceSet(v.Elem(), x(1), option{})
	if err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatal(i)
	}
}

func TestForceSetAlias2(t *testing.T) {
	var i x = 0
	v := reflect.ValueOf(&i)

	err := ForceSet(v.Elem(), 1, option{})
	if err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatal(i)
	}
}

type User struct {
	Name string
}

type Person User

func TestForceSetStruct(t *testing.T) {
	var i User
	v := reflect.ValueOf(&i)

	err := ForceSet(v.Elem(), Person{
		Name: "fun",
	}, option{})
	if err != nil {
		t.Fatal(err)
	}
	if i.Name != "fun" {
		t.Fatal(i)
	}
}

func TestForceSetInterface(t *testing.T) {
	var i RoleInfo
	v := reflect.ValueOf(&i)
	name := v.Elem().FieldByName("Role")
	err := ForceSet(name, &Admin{}, option{})
	if err != nil {
		t.Fatal(err)
	}
	if i.Role.Name() != "admin" {
		t.Fatal(i.Role.Name())
	}
}

func TestForceSetString(t *testing.T) {
	var str string
	v := reflect.ValueOf(&str).Elem()
	err := ForceSet(v, 12, option{})
	if err != nil {
		t.Fatal(err)
	}
	if str != "12" {
		t.Fatal(str)
	}
}

type RoleInfo struct {
	Role Role
}

type Role interface {
	Name() string
}

type Admin struct{}

func (a *Admin) Name() string { return "admin" }
