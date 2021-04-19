package forceset

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

type x int

func TestForceSetAlias(t *testing.T) {
	var i int = 0
	v := reflect.ValueOf(&i)
	err := ForceSet(v.Elem(), x(1))
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

	err := ForceSet(v.Elem(), 1)
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
	})
	if err != nil {
		t.Fatal(err)
	}
	if i.Name != "fun" {
		t.Fatal(i)
	}
}

func TestForceSetFromJSON(t *testing.T) {
	var i User
	v := reflect.ValueOf(&i)

	err := ForceSet(v.Elem(), json.RawMessage(`{"Name":"Peter"}`))
	if err != nil {
		t.Fatal(err)
	}
	if i.Name != "Peter" {
		t.Fatal(i)
	}
	m := map[string]interface{}{}
	Set(&m, json.RawMessage(`{"Name":"Peter"}`))
	if m["Name"] != "Peter" {
		t.Fatal(m)
	}
}

func TestForceSetInterface(t *testing.T) {
	var i RoleInfo
	v := reflect.ValueOf(&i)
	name := v.Elem().FieldByName("Role")
	err := ForceSet(name, &Admin{})
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
	err := ForceSet(v, 12)
	if err != nil {
		t.Fatal(err)
	}
	if str != "12" {
		t.Fatal(str)
	}
}

type Address struct {
	Code float64
	Text ***string `json:"TEXT"`
}

type Address2 struct {
	Code int
	Text string
}

func TestSetStructFromMap(t *testing.T) {
	obj := map[string]interface{}{
		"TEXT": 1,
		"Code": "2",
	}
	var s = &Address{}
	err := Set(s, obj)
	if err != nil {
		t.Fatal(err)
	}
	if ***s.Text != "1" || s.Code != 2 {
		t.Fatalf("%#v", s)
	}
}

type Address3 struct {
	*Address
}

func TestSetStructFromMap2(t *testing.T) {
	obj := map[string]interface{}{}
	var s = &Address3{}
	err := Set(s, obj)
	if err != nil {
		t.Fatal(err)
	}
	if s.Address != nil {
		t.Fatal("should be nil")
	}
}

type Address4 struct {
	*Address
	Name string
}

func TestSetStructFromMap3(t *testing.T) {
	obj := map[string]interface{}{
		"TEXT": 1,
		"Code": "2",
		"Name": "name",
	}
	var s = &Address4{}
	err := Set(s, obj)
	if err != nil {
		t.Fatal(err)
	}
	if s.Address == nil || ***s.Text != "1" || s.Code != 2 || s.Name != "name" {
		t.Fatalf("%#v", s)
	}
}

func TestSetMapFromStruct(t *testing.T) {
	obj := map[string]interface{}{}
	var a2 = &Address2{
		Code: 2,
		Text: "1",
	}
	var a *Address
	err := Set(&a, a2)
	if err != nil {
		t.Fatal(err)
	}
	err = Set(&obj, a)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", obj)
	if ***(obj["TEXT"].(***string)) != "1" {
		t.Fatal("expected TEXT == 1 got:", obj["TEXT"])
	}
	if obj["Code"] != float64(2) {
		t.Fatal("expected Code == 2 got:", obj["Code"])
	}
}

func TestSetStructFromStruct(t *testing.T) {
	obj := &Address2{
		Code: 32,
		Text: "text",
	}
	var s = &Address{}
	err := Set(s, obj)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", s)
	t.Log(***s.Text)
}

func TestSetSliceFromMap(t *testing.T) {
	m := map[int]string{1: "2", 3: "4"}
	var l []int
	err := Set(&l, m, func(opt *SetOption) {
		opt.MapToSliceOption = ArrayLike
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l, []int{0, 2, 0, 4}) {
		t.Fatal("expected []int{0, 2, 0, 4} got:", l)
	}
}

func TestSetSliceFromMap2(t *testing.T) {
	type pair struct {
		Key   string
		Value string
	}
	m := map[int]string{1: "2", 3: "4"}
	var l []pair
	err := Set(&l, m, func(opt *SetOption) {
		opt.MapToSliceOption = Pairs
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l, []pair{{"1", "2"}, {"3", "4"}}) {
		t.Fatal("expected []pair{{\"1\", \"2\"}, {\"3\", \"4\"}} got:", l)
	}
}

func TestSetSliceFromStruct(t *testing.T) {
	var slice []string
	var addr = Address2{
		Code: 1,
		Text: "sometext",
	}
	err := Set(&slice, addr)
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"1", "sometext"}
	if !reflect.DeepEqual(expected, slice) {
		t.Fatal(slice)
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

func TestSetTime(t *testing.T) {
	type Data struct {
		Time *time.Time `json:"time;format:2006-01-02 15:04:05"`
		Next *time.Time
	}
	var expect = "2020-05-19 16:20:17"
	var format = "2006-01-02 15:04:05"

	var data = map[string]interface{}{
		"time": expect,
	}
	var d Data
	err := Set(&d, data, func(opt *SetOption) {
		opt.Tag = "json"
		opt.Mappers[MapperType{
			Destination: reflect.TypeOf(time.Time{}),
			Source:      reflect.TypeOf(""),
		}] = func(dst reflect.Value, src reflect.Value, tag string) error {
			lines := strings.Split(tag, ";")
			var fmt string
			for _, line := range lines {
				if strings.HasPrefix(line, "format:") {
					fmt = line[len("format:"):]
					break
				}
			}
			if fmt == "" {
				fmt = time.RFC3339
			}
			t.Log(fmt)
			parse, err := time.Parse(fmt, src.Interface().(string))
			if err != nil {
				return err
			}
			dst.Set(reflect.ValueOf(parse))
			return nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Time == nil || d.Time.Format(format) != expect {
		t.Fatal("expected :", expect, "got:", d.Time.Format(format))
	}
	if d.Next != nil {
		t.Fatal("expected nil got:", d.Next)
	}
}
