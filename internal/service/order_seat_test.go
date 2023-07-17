package service

import (
	"fmt"
	"math/rand"
	"seats/internal/models"
	"testing"
	"time"
)

// 初始化票档
var ticketLevels = map[models.Level]string{
	models.Level0: "外场288",
	models.Level1: "外场488",
	models.Level2: "外场688",
	models.Level3: "内场888",
	models.Level4: "内场1088",
	models.Level5: "内场1288",
}

// 从主办方获知了真实票源信息，知道了每个档位对应的座位区
var areaLevels = map[int]models.Level{
	// 外场188
	0: models.Level0,
	1: models.Level0,
	// 外场388
	10: models.Level1,
	11: models.Level1,
	12: models.Level1,
	15: models.Level1,
	16: models.Level1,
	17: models.Level1,
	18: models.Level1,
	19: models.Level1,
	// 外场588
	21: models.Level2,
	23: models.Level2,
	25: models.Level2,
	27: models.Level2,
	29: models.Level2,
	// 内场688
	33: models.Level3,
	34: models.Level3,
	35: models.Level3,
	36: models.Level3,
	37: models.Level3,
	38: models.Level3,
	39: models.Level3,
	// 内场988
	40: models.Level4,
	42: models.Level4,
	44: models.Level4,
	46: models.Level4,
	48: models.Level4,
	// 内场1288
	50: models.Level5,
	51: models.Level5,
	52: models.Level5,
	53: models.Level5,
	54: models.Level5,
}

const (
	// 每一个座位区每一排最多有多少个座位
	maxSeatsPerLine = 60
	// 每一个座位区多少排位置
	maxLinesPerArea = 60
)

func randomSeatsInLine(max int) map[int]bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var seatIDs = make(map[int]bool)
	// 随机生成 max 个座位号
	// 对应每个座位是否锁定 1/10 概率锁定
	for i := 0; i < max; i++ {
		id := r.Int31n(int32(max))
		lockNum := r.Int31n(10)
		seatIDs[int(id)] = lockNum == 0
	}

	return seatIDs
}

// 初始化座位
func initSeats() (*models.Event, map[bool]int) {
	fmt.Printf("初始化全场座位信息 ...\n")
	var event = models.NewEvent(0, "测试活动", time.Now())

	// 初始化档位信息
	for k, v := range ticketLevels {
		event.AddLevel(k, v)
	}

	// 初始化座位区
	for k, v := range areaLevels {
		event.AddArea(k, v)
	}

	// 座位区情况随机生成
	// 保证有间断的座位和不间断的座位
	// 保证有锁定的座位和没锁定的座位

	var totalSeats = make(map[bool]int)
	// 为每一个座位区初始化座位
	for area := range areaLevels {
		// 为每一行座位初始化座位
		for line := 1; line <= maxLinesPerArea; line++ {
			// 生成
			// 随机生成这一行可用的所有座位（每一个座位附带是否被锁定了的信息）
			seatIDs := randomSeatsInLine(maxSeatsPerLine)
			var lineTotal = make(map[bool]int)
			for id, locked := range seatIDs {
				totalSeats[locked]++
				lineTotal[locked]++
				seat := models.NewSeat(area, line, id, models.SeatAvailible)
				// 生成座位
				if locked {
					seat.SetStatus(models.SeatLocked)
				}

				// 把座位信息加进去
				event.Add(seat)

			}
			// fmt.Printf("%d 生成 %d 个座位，可用: %d, 不可用: %d: %v\n", line, len(seatIDs), lineTotal[false], lineTotal[true], seatIDs)
		}
	}

	// 返回所有可用的连坐信息（一个座位的也在里面）
	return event, totalSeats
}

// 模拟出票
// 随机选择一定数量的用户来出票
// total: 总出票数量
// 最大数量：20
func randomOrderSeat(total map[models.Level]int) map[models.Level][]int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var info = make(map[models.Level][]int)
	var max = 20

	for level, t := range total {
		for t > 0 {
			if t < max {
				max = t
			}
			// 随机生成n个用户来出票
			// 最多20人
			n := int(r.Int31n(int32(max)) + 1)
			t -= n

			info[level] = append(info[level], n)
		}
	}

	return info
}

func TestOrderSeat(t *testing.T) {
	// 初始化所有可用座位信息
	event, seats := initSeats()
	event.Info()
	fmt.Printf("total seats: %v\n", seats)

	// 生成随机出票信息
	ticketInfo := randomOrderSeat(event.GetLevelCount())
	var totalUsersToPick int
	var usersToPick = make(map[models.Level]int)
	for level, ns := range ticketInfo {
		for _, n := range ns {
			usersToPick[level] += n
			totalUsersToPick += n
		}
	}
	for level, count := range usersToPick {
		fmt.Printf("========== %s档位%d名用户等待出票\n", event.GetLevelName(level), count)
	}
	fmt.Printf("========== 总计%d名用户等待出票\n", totalUsersToPick)

	var totalOrderedUsers int
	var orderedUsers = make(map[models.Level]int)
	var usersWaiting = make(map[models.Level]int)
	for level, ns := range ticketInfo {
		for _, n := range ns {
			// level: 票档
			// n: 出票数量
			fmt.Printf("********** 连坐求解：%s(%d)档位%d人连坐。剩余座位数: %d, 剩余待出票人数: %d\n", event.GetLevelName(level), event.FindArea(level), n, event.CountTotal(event.FindArea(level)...), usersToPick[level]-orderedUsers[level])

			// 尝试在每一个区里面找到最优的连坐
			fullyPicked, partialPicked := event.PickBest(n, event.FindArea(level)...)

			// 如果没有找到合适的连坐
			if fullyPicked == nil && partialPicked == nil {
				fmt.Printf("======= %s(%d)档位%d人连坐 未找到合适的连坐\n\n", event.GetLevelName(level), event.FindArea(level), n)
				event.Info()
				for level, count := range orderedUsers {
					fmt.Printf("++++++ %s 档位 %d 名用户已出票\n", event.GetLevelName(level), count)
				}
				fmt.Printf("++++++ 总计 %d 名用户已出票\n", totalOrderedUsers)
				return
			}

			fmt.Printf("----- %s(%d)档位%d人连坐 找到合适的连坐，准备订座 ...\n", event.GetLevelName(level), event.FindArea(level), n)
			for _, p := range fullyPicked {
				fmt.Printf("+++++ 完全使用 %s ...\n", p.String())
			}

			for k, v := range partialPicked {
				fmt.Printf("+++++ 部分使用(%d) %s ...\n", v, k.String())
			}

			for _, picked := range fullyPicked {
				orderedSeats, ok := picked.Order(picked.CountTotal())
				if !ok {
					usersWaiting[level] += n
					fmt.Printf("======= %s(%d) 档位 %d 人连坐 订座失败: %s\n\n", event.GetLevelName(level), event.FindArea(level), n, picked)
					return
				}
				fmt.Printf("======= %s(%d) 档位 %d 人连坐 订座信息：【%s】\n", event.GetLevelName(level), event.FindArea(level), n, orderedSeats.String())
			}

			for seat, n := range partialPicked {
				orderedSeats, ok := seat.Order(n)
				if !ok {
					usersWaiting[level] += n
					fmt.Printf("======= %s(%d) 档位 %d 人连坐 订座失败: %s\n\n", event.GetLevelName(level), event.FindArea(level), n, seat)
					return
				}
				fmt.Printf("======= %s(%d) 档位 %d 人连坐 订座信息：【%s】\n", event.GetLevelName(level), event.FindArea(level), n, orderedSeats.String())
			}

			println()

			orderedUsers[level] += n
			totalOrderedUsers += n

			// dao 更新已订座的座位信息
			// ...
			// ...
		}
	}
	event.Info()

	for level, count := range orderedUsers {
		fmt.Printf("++++++ %s 档位 %d 名用户已出票\n", event.GetLevelName(level), count)
	}
	fmt.Printf("++++++ 总计 %d 名用户已出票\n", totalOrderedUsers)

}
