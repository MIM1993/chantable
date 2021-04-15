package test

import (
	"fmt"
	"testing"
)

/*
	copy(arr[i:], arr[i+1:])
	arr = arr[:len(arr)-1]
*/
func TestArrCopy(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Println(arr)
	i := 9
	copy(arr[i:], arr[i+1:])
	arr = arr[:len(arr)-1]

	fmt.Println(arr)
}
