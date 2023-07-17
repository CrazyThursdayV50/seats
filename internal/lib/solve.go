package lib

import (
	"sort"
)

// by:
// key: num
// value: count
// 从 by 中找到能加起来等于 num 的结果
// 或者从 by 中找到能加起来略大于 num 的结果，并附带“大了多少”
// 作用：在选连坐的时候，如果一个大连坐需求无法满足，则会将大的连坐需求转化成多个小的连坐来求解
// 并且在还要满足一种特殊场景，例如：场内一共有8个座位，为4个2连坐，此时需要找一个7连坐，那么应该将其中3个2连坐订座，并且还有一个2连坐订下一个位置。
func Solve(num int, by map[int]int) (map[int]int, map[int]int) {
	if by == nil {
		return nil, nil
	}

	_, ok := by[num]
	if ok {
		return map[int]int{num: 1}, nil
	}

	var bySlice []int
	for k := range by {
		bySlice = append(bySlice, k)
	}

	// 从大到小排
	sort.Sort(sort.Reverse(sort.IntSlice(bySlice)))

	var resultMap = make(map[int]int)
	var left = num
	for _, n := range bySlice {
		// 如果存在一个大于 left 的数，说明可以满足
		if n > left {
			// 看存不存在恰好相等的
			_, ok := by[left]
			if !ok {
				resultMap[n]++
				return resultMap, map[int]int{n: left}
			}

			resultMap[left]++
			return resultMap, nil
		}

		// 如果相等
		if n == left {
			resultMap[n]++
			return resultMap, nil
		}

		// 如果 n < left
		totalN := n * by[n]
		// 如果刚好满足
		if totalN == left {
			resultMap[n] += by[n]
			return resultMap, nil
		}

		// 如果总数小于 left
		if totalN < left {
			resultMap[n] += by[n]
			left -= totalN
			continue
		}

		// 如果总数大于 left
		// 看看 left 和 n 的倍数关系
		times := left / n
		last := left % n

		if last == 0 {
			resultMap[n] += times
			return resultMap, nil
		}

		resultMap[n] += times + 1
		return resultMap, map[int]int{n: left % n}
	}

	// 无解
	return nil, nil
}
