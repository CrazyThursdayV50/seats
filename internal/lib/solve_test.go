package lib

import (
	"fmt"
	"testing"
)

func TestSolve(t *testing.T) {
	var by = map[int]int{
		1:  1,
		2:  1,
		3:  1,
		4:  1,
		5:  1,
		6:  1,
		7:  1,
		8:  1,
		9:  1,
		10: 1,
		11: 1,
		12: 1,
	}
	var dst = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	for _, num := range dst {
		fmt.Printf("solve num %d\n", num)
		full, part := Solve(num, by)
		fmt.Printf("dst: %d, full: %v, part: %v\n", num, full, part)
	}
}
