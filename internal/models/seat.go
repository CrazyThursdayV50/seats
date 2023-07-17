package models

import "fmt"

type SeatStatus int

const (
	SeatAvailible SeatStatus = iota
	// 座位不可用
	SeatUnavalible
	// 在出票过程中被暂时预选了
	SeatPicked
	// 已出票，座位被锁定
	SeatOrdered
	// 座位被征用，已被锁定
	SeatLocked
)

// 虚拟座位号
// 一共12位
// 状态 区   排  列
// 99  9999 999 999
type SeatID int64

const (
	SeatLineMul   = 1e3
	SeatAreaMul   = 1e3 * SeatLineMul
	SeatStatusMul = 1e4 * SeatAreaMul

	MaxColumn = 999
	MaxLine   = 999 * SeatLineMul
	MaxArea   = 9999 * SeatAreaMul
)

func NewSeatID(status SeatStatus, area, line, column int) SeatID {
	return SeatID(int64(column) + int64(line)*SeatLineMul + int64(area)*SeatAreaMul + int64(status)*SeatStatusMul)
}

func (s SeatID) String() string {
	return fmt.Sprintf("%04d区%03d排%03d列", s.GetArea(), s.GetLine(), s.GetColumn())
}

func (s SeatID) GetStatus() SeatStatus {
	return SeatStatus(s / SeatStatusMul)
}

func (s SeatID) removeStatus() int64 {
	return int64(s) % SeatStatusMul
}

func (s SeatID) GetArea() int {
	return int(s.removeStatus() / SeatAreaMul)
}

func (s SeatID) GetLine() int {
	return int((s % SeatAreaMul) / SeatLineMul)
}

func (s SeatID) GetColumn() int {
	return int(s % SeatLineMul)
}

type Seat struct {
	SeatID
}

func (s *Seat) SetStatus(status SeatStatus) {
	s.SeatID = SeatID(s.removeStatus() + int64(status*SeatStatusMul))
}

func (s *Seat) Order() {
	s.SetStatus(SeatOrdered)
}

// 新创建座位
func NewSeat(area, line, column int, status SeatStatus) *Seat {
	return &Seat{SeatID: NewSeatID(status, area, line, column)}
}

// 座位是否可用
func (s *Seat) IsAvailible() bool { return s.GetStatus() == SeatAvailible }

// 某区域可用的座位号最小值
func AreaAvailibleMin(area int) SeatID {
	return SeatID(int(SeatAvailible)*SeatStatusMul + area*SeatAreaMul)
}

// 某区域可用的座位号最大值
func AreaAvailibleMax(area int) SeatID {
	return SeatID(int(SeatAvailible)*SeatStatusMul + area*SeatAreaMul + MaxLine + MaxColumn)
}

func seatStatusMin(status SeatStatus) SeatID {
	return SeatID(int(status) * SeatStatusMul)
}

func seatStatusMax(status SeatStatus) SeatID {
	return SeatID(int(status)*SeatStatusMul + MaxArea + MaxLine + MaxColumn)
}

// 可用的座位号最小值
func AvailibleMin() SeatID {
	return seatStatusMin(SeatAvailible)
}

// 可用的座位号最大值
func AvailibleMax() SeatID {
	return seatStatusMax(SeatAvailible)
}

// 已订的座位号最小值
func OrderedMin() SeatID {
	return seatStatusMin(SeatOrdered)
}

// 已订的座位号最大值
func OrderedMax() SeatID {
	return seatStatusMax(SeatOrdered)
}

// 锁定的座位号最小值
func LockedMin() SeatID {
	return seatStatusMin(SeatLocked)
}

// 锁定的座位号最大值
func LockedMax() SeatID {
	return seatStatusMax(SeatLocked)
}
