package main

import (
	"fmt"
	"math"
	"strconv"
)

type Int struct {
	i int64
}

func structFieldPointer() {
	var v = Int{i: 10}

	fmt.Printf("%T %v\n", &v.i, &v.i)
	fmt.Printf("%T %v\n", &(v.i), &(v.i))
	fmt.Printf("%T %v\n", (&v).i, (&v).i)
	fmt.Printf("%T %v\n", &v, &v)
}

func floatUint() {
	res_1 := math.Float64frombits(2)
	res_2 := math.Float64frombits(1)
	res_3 := math.Float64frombits(0)
	res_4 := math.Float64frombits(23)

	// Displaying the result
	fmt.Println("Result 1: ", res_1)
	fmt.Println("Result 2: ", res_2)
	fmt.Println("Result 3: ", res_3)
	fmt.Println("Result 4: ", res_4)

	v := 3.1415926535

	s32 := strconv.FormatFloat(v, 'f', 4, 32)
	s64 := strconv.FormatFloat(v, 'E', 6, 64)
	fmt.Printf("%T, %v\n", s32, s32)
	fmt.Printf("%T, %v\n", s64, s64)
}

func jsonMarshal() {
	s := "aaa"
	//v := json.Marshal(s)
	fmt.Printf("%T %v", s, s)
}

func main() {
	jsonMarshal()
}
