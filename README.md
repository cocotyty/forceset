ForceSet
---
ForceSet provides an easier way to set the value of a variable. 
It prevents you from writing a lot of type conversion code.
## Example
```go
package main

import (
	"fmt"
	"github.com/cocotyty/forceset"
)

// Pair is map's kv pair.
type Pair struct {
	Key   string // first field represents key ,  must be exported .
	Label string // second field represents value ,  must be exported .
}

func main() {
	// basic type convert
	var i int32
	_ = forceset.Set(&i, "32")
	fmt.Printf("convert %#v to %#v \n", "32", i)

	var b bool
	_ = forceset.Set(&b, "True")
	fmt.Printf("convert %#v to %#v \n", "True", b)

	// slice to slice
	var arr []int
	_ = forceset.Set(&arr, []string{"1", "2"})
	fmt.Printf("convert %#v to %#v \n", []string{"1", "2"}, arr)

	// slice to slice
	var args []interface{}
	_ = forceset.Set(&args, []string{"1", "2"})
	fmt.Printf("convert %#v to %#v \n", []string{"1", "2"}, args)

	// map to map
	var m2 = map[string]string{}
	var m = map[int]int{1: 2}
	_ = forceset.Set(&m2, m)
	fmt.Printf("convert %#v to %#v \n", map[int]int{1: 2}, m2)

	// map to slice
	options := map[string]string{
		"#F00": "Red",
		"#0F0": "Green",
		"#00F": "Blue",
	}
	var optionList []Pair
	_ = forceset.Set(&optionList, options, forceset.MapAsPairs)
	fmt.Printf("convert %#v to %#v \n", options, optionList)

	// map to slice
	phpArray := map[string]string{"0": "Red", "1": "Green", "2": "Blue"}
	var goArray []string
	_ = forceset.Set(&goArray, phpArray, forceset.MapAsArrayLike)

	fmt.Printf("convert %#v to %#v \n", phpArray, goArray)
}
```