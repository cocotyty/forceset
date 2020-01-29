package main

import (
	"fmt"
	"github.com/cocotyty/forceset"
)

type Pair struct {
	Key   string
	Label string
}

func main() {

	var i int32
	_ = forceset.Set(&i, "32")
	fmt.Printf("convert %#v to %#v \n", "32", i)

	var b bool
	_ = forceset.Set(&b, "True")
	fmt.Printf("convert %#v to %#v \n", "True", b)

	var arr []int
	_ = forceset.Set(&arr, []string{"1", "2"})
	fmt.Printf("convert %#v to %#v \n", []string{"1", "2"}, arr)

	var args []interface{}
	_ = forceset.Set(&args, []string{"1", "2"})
	fmt.Printf("convert %#v to %#v \n", []string{"1", "2"}, args)

	var m2 = map[string]string{}
	var m = map[int]int{1: 2}
	_ = forceset.Set(&m2, m)
	fmt.Printf("convert %#v to %#v \n", map[int]int{1: 2}, m2)

	options := map[string]string{
		"#F00": "Red",
		"#0F0": "Green",
		"#00F": "Blue",
	}
	var optionList []Pair
	_ = forceset.Set(&optionList, options, forceset.MapAsPairs)
	fmt.Printf("convert %#v to %#v \n", options, optionList)

	phpArray := map[string]string{"0": "Red", "1": "Green", "2": "Blue"}
	var goArray []string
	_ = forceset.Set(&goArray, phpArray, forceset.MapAsArrayLike)

	fmt.Printf("convert %#v to %#v \n", phpArray, goArray)

}
