package main

import "fmt"

func main() {
	var randomSlice []string

	randomSlice = append(randomSlice, "haha!")

	fmt.Printf("0: [%v]", randomSlice[0])
	fmt.Printf("1: [%v]", randomSlice[1])

}
