package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	qwe(&wg)
	fmt.Println("after qwe()")
}

func qwe(wg *sync.WaitGroup) {
	defer func() {
		//wg.Wait()
		time.Sleep(5 * time.Second)
		fmt.Println("after wait")
	}()
	return
}
