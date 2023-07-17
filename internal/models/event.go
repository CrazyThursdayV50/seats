package models

import (
	"time"
)

// 档位
type Level int

const (
	Level0 Level = 0
	Level1 Level = 1
	Level2 Level = 2
	Level3 Level = 3
	Level4 Level = 4
	Level5 Level = 5
	Level6 Level = 6
	Level7 Level = 7
	Level8 Level = 8
	Level9 Level = 9
)

// 活动场次
type Event struct {
	// 活动唯一id
	id int
	// 活动时间
	datetime time.Time
	// 活动名称
	name string
	// 档位信息
	// key: level
	// value: level name
	levels map[Level]string
	// 座位区与档位的对应情况
	// 每一个档位对应哪些座位区
	// key: area id
	// value: level
	areaLevels map[int]Level
	// 每一个座位区对应哪个票档
	// key: level
	// value: []area_id
	levelAreas map[Level][]int

	Seats
}

func NewEvent(id int, name string, datetime time.Time) *Event {
	return &Event{id: id, datetime: datetime, name: name, levels: make(map[Level]string), areaLevels: make(map[int]Level), levelAreas: make(map[Level][]int)}
}

// 增加档位信息
func (e *Event) AddLevel(lvl Level, name string) {
	e.levels[lvl] = name
}

// 增加座位区情况
func (e *Event) AddArea(area int, lvl Level) {
	if _, ok := e.areaLevels[area]; ok {
		return
	}
	e.areaLevels[area] = lvl
	e.levelAreas[lvl] = append(e.levelAreas[lvl], area)
}

// 查询
func (e *Event) FindArea(lvl Level) []int {
	return e.levelAreas[lvl]
}

func (e *Event) FindLevel(area int) Level {
	return e.areaLevels[area]
}

func (e *Event) GetLevelName(lvl Level) string {
	return e.levels[lvl]
}

// 获取档位对应的座位数量
func (e *Event) GetLevelCount() map[Level]int {
	var result = make(map[Level]int)
	for level, areas := range e.levelAreas {
		result[level] += e.CountTotal(areas...)
	}

	return result
}

func (e *Event) Info() {
	println(e.Seats.Info())
}
