package main

import (
	"fmt"
	"runtime"
)

func main() {
	//slice := make([]int, 0)  // []
	//slice = append(slice, 1) // [1]
	//slice = append(slice, 2) // [1, 2]
	//slice = append(slice, 3) //
	//
	//a := slice
	//b := slice
	//
	//a = append(a, 5)
	//b = append(b, 6)
	fmt.Println("numcpu", runtime.NumCPU())
	var m = make(map[int]int)
	//time.Sleep(time.Second * 60)
	//m[1] = 5
	fmt.Println("second", m[5], &m)
	println(&m)
	//
	//arr := [5]int{0, 1, 2, 3, 4}
	//
	//s := arr[:3]
	//s2 := append(s, 99, 100, 101)
	//fmt.Println(arr, s, s2)
}
