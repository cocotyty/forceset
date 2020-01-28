package main

import (
	"fmt"
	"github.com/cocotyty/forceset"
	"reflect"
)

func main() {

	var i int32
	_ = forceset.ForceSet(reflect.ValueOf(&i).Elem(), "32")

	var b bool
	_ = forceset.ForceSet(reflect.ValueOf(&b).Elem(), "True")

	var arr []int
	_ = forceset.ForceSet(reflect.ValueOf(&arr).Elem(), []string{"1", "2"})

	var args []interface{}
	_ = forceset.ForceSet(reflect.ValueOf(&args).Elem(), []string{"1", "2"})
	fmt.Println(i, b, arr, args)
}
