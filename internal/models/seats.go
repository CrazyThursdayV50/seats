package models

import (
	"fmt"
	"seats/internal/lib"
	"sort"
	"strings"
)

type Seats []*Seat

func (s Seats) Len() int           { return len(s) }
func (s Seats) LastIndex() int     { return s.Len() - 1 }
func (s Seats) Less(i, j int) bool { return s[i].SeatID < s[j].SeatID }
func (s Seats) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *Seats) Add(seat *Seat)    { *s = append(*s, seat) }

func (s Seats) String() string {
	return fmt.Sprintf("从 %s 到 %s, 共 %d座位", s[0].String(), s[s.LastIndex()].String(), s.Len())
}

// 总数量：
// 1. 可用数量
// 2. 锁定数量
// 3. 已定数量
// 每个区座位数量
func (s Seats) Info() string {
	seatInfoFormat := "总数[%d], 可用[%d], 已出[%d], 锁定[%d]"
	content := []string{fmt.Sprintf("总数据: "+seatInfoFormat, s.CountTotal(), s.CountAvailable(), s.CountOrdered(), s.CountLocked())}
	areas, areaSeats := s.GroupByArea()
	for _, area := range areas {
		seats := areaSeats[area]
		content = append(content, fmt.Sprintf("%d区: "+seatInfoFormat, area, seats.CountTotal(), seats.CountAvailable(), seats.CountOrdered(), seats.CountLocked()))
	}

	return strings.Join(content, "\n")
}

// 整理座位
func (s Seats) tidy() {
	sort.Slice(s, func(i, j int) bool {
		si := s[i]
		sj := s[j]
		return si.SeatID < sj.SeatID
	})
}

// 找到区域内可用的座位
func (s Seats) FindAvailableInAreas(areas ...int) map[int]Seats {
	var result = make(map[int]Seats)
	for _, area := range areas {
		from := AreaAvailibleMin(area)
		to := AreaAvailibleMax(area)
		for _, seat := range s {
			if seat.SeatID >= from && seat.SeatID <= to {
				result[area] = append(result[area], seat)
			}
		}
		sort.Sort(result[area])
	}

	return result
}

func (s Seats) copy() Seats {
	var seats = make(Seats, s.Len(), s.Len())
	copy(seats, s)
	return seats
}

// 生成连坐
func (s Seats) GenConsecutiveSeats() []Seats {
	// 座位号小的排前面
	sort.Sort(s)
	var result = make([]Seats, 0)
	// 连号即连坐
	var lastID int

	var tempSeats = make(Seats, 0)
	for i, seat := range s {
		// 如果连坐没有信息，直接填充进去
		if tempSeats.Len() == 0 {
			tempSeats.Add(seat)
			lastID = seat.GetColumn()
			if i == s.LastIndex() {
				result = append(result, tempSeats.copy())
			}
			continue
		}

		// 如果连号，则加进去
		if seat.GetColumn() == lastID+1 {
			tempSeats.Add(seat)
			lastID = seat.GetColumn()
			if i == s.LastIndex() {
				result = append(result, tempSeats.copy())
			}
			continue
		}

		// 如果不是连号
		// 连坐中断
		result = append(result, tempSeats.copy())

		// 本座位为下一个连坐的开始
		tempSeats = []*Seat{seat}
		lastID = seat.GetColumn()
		if i == s.LastIndex() {
			result = append(result, tempSeats.copy())
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Len() < result[j].Len()
	})

	return result
}

type areaSeats struct {
	area  int
	seats Seats
}

func (s Seats) PickBest(num int, areas ...int) ([]Seats, map[*Seats]int) {
	sort.Ints(areas)

	// 1. 生成区域内连坐
	areasSeats := s.FindAvailableInAreas(areas...)

	// 分别在各自区域中生成连坐
	availibleSeats := make([]*areaSeats, 0)
	// 区域连坐信息
	var areaConSeats = make(map[int][]Seats)
	for area, seats := range areasSeats {
		// 在区域内生成连坐
		conSeats := seats.GenConsecutiveSeats()
		areaConSeats[area] = conSeats
		// 2. 从连坐中找到最匹配的
		for _, s := range conSeats {
			// 2.1.1 找到了，结束
			if s.Len() >= num {
				availibleSeats = append(availibleSeats, &areaSeats{area, s})
				break
			}
		}
	}

	// 2.1.1 如果找到了
	if len(availibleSeats) != 0 {
		// 2.1.1.1把区号小的排前面
		sort.Slice(availibleSeats, func(i, j int) bool {
			return availibleSeats[i].area < availibleSeats[j].area
		})

		seats := availibleSeats[0].seats
		if seats.CountTotal() == num {
			return []Seats{seats}, nil
		}
		// 2.1.1.2 返回第一个可用的
		return nil, map[*Seats]int{&seats: num}
	}

	// 2.1.2 如果没有找到
	// 尝试将大连坐分成小连坐去满足需求

	// 2.1.2.1 优先在同一区域满足需求
	var totalConSeats []Seats
	for _, area := range areas {
		conSeats := areaConSeats[area]
		var src = make(map[int]int)
		var seatGroup = make(map[int][]Seats)
		for _, conSeat := range conSeats {
			src[conSeat.Len()]++
			// 按照连坐长度分组
			seatGroup[conSeat.Len()] = append(seatGroup[conSeat.Len()], conSeat)
			totalConSeats = append(totalConSeats, conSeat)
		}

		// 求解
		full, part := lib.Solve(num, src)

		// 有解
		if full != nil {
			var result []Seats
			// seatLen: 连坐长度
			// seatCount: 连坐个数
			for seatLen, seatCount := range full {
				if _, ok := part[seatLen]; ok {
					result = append(result, seatGroup[seatLen][1:seatCount]...)
					continue
				}
				result = append(result, seatGroup[seatLen][:seatCount]...)
			}

			if part != nil {
				var partMap = make(map[*Seats]int)
				for k, v := range part {
					partMap[&seatGroup[k][0]] = v
				}

				return result, partMap
			}

			return result, nil
		}

	}

	// 2.1.2.2 同一区域无解时，在所有区域中求解
	var src = make(map[int]int)
	var seatGroup = make(map[int][]Seats)
	for _, conSeats := range totalConSeats {
		src[conSeats.Len()]++
		seatGroup[conSeats.Len()] = append(seatGroup[conSeats.Len()], conSeats)
	}

	full, part := lib.Solve(num, src)

	// 有解
	if full != nil {
		var result []Seats
		// seatLen: 连坐长度
		// seatCount: 连坐个数
		for seatLen, seatCount := range full {
			if _, ok := part[seatLen]; ok {
				result = append(result, seatGroup[seatLen][1:seatCount]...)
				continue
			}
			result = append(result, seatGroup[seatLen][:seatCount]...)
		}

		if part != nil {
			var partMap = make(map[*Seats]int)
			for k, v := range part {
				partMap[&seatGroup[k][0]] = v
			}

			return result, partMap
		}

		return result, nil
	}

	return nil, nil
}

// 连坐订座
func (s Seats) Order(num int) (Seats, bool) {
	if s.Len() < num {
		return nil, false
	}

	seats := make(Seats, 0, num)
	for _, seat := range s[:num] {
		seat.Order()
		seats.Add(seat)
	}

	s.tidy()
	return seats, true
}

// 查询座位数
func (s Seats) CountTotal(areas ...int) int {
	if areas == nil {
		return s.Len()
	}

	areaSeats := s.FindAvailableInAreas(areas...)
	var total int
	for _, v := range areaSeats {
		total += v.Len()
	}
	return total
}

func (s Seats) CountAvailable() int {
	from := AvailibleMin()
	to := AvailibleMax()
	var count int
	for _, seat := range s {
		if seat.SeatID >= from && seat.SeatID <= to {
			count++
		}
	}

	return count
}

func (s Seats) CountOrdered() int {
	from := OrderedMin()
	to := OrderedMax()
	var count int
	for _, seat := range s {
		if seat.SeatID >= from && seat.SeatID <= to {
			count++
		}
	}

	return count
}

func (s Seats) CountLocked() int {
	from := LockedMin()
	to := LockedMax()
	var count int
	for _, seat := range s {
		if seat.SeatID >= from && seat.SeatID <= to {
			count++
		}
	}

	return count
}

func (s Seats) GroupByArea() ([]int, map[int]Seats) {
	var group = make(map[int]Seats)
	for _, seat := range s {
		area := seat.GetArea()
		group[area] = append(group[area], seat)
	}

	var areas []int
	for k := range group {
		areas = append(areas, k)
	}

	sort.Ints(areas)
	return areas, group
}
