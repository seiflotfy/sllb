package shll

import (
	"errors"
	"math"

	metro "github.com/dgryski/go-metro"
)

var (
	exp32 = math.Pow(2, 32)
)

/*
SlidingHyperLogLog adapts the HyperLogLog algorithm of Flajolet et. al to the
data stream processing by adding a sliding window mechanism. It has the
advantage to estimate at any time the number of flows seen over any duration
bounded by the length of the sliding window.
*/
type SlidingHyperLogLog struct {
	window          uint64
	alpha           float64
	p               uint
	m               uint
	regs            []*reg
	latestTimestamp uint64
}

/*
NewSlidingHyperLogLog return a new SlidingHyperLogLog.
errorRate = abs_err / cardinality
window = window to keep track of
*/
func NewSlidingHyperLogLog(errRate float64) (*SlidingHyperLogLog, error) {
	if !(0 < errRate && errRate < 1) {
		return nil, errors.New("errRate must be between 0 and 1")
	}
	p := uint(math.Ceil(math.Log2(math.Pow((1.04 / errRate), 2))))
	m := uint(1) << p
	shll := &SlidingHyperLogLog{
		p:     p,
		m:     m,
		regs:  make([]*reg, m, m),
		alpha: alpha(float64(m)),
	}
	for i := range shll.regs {
		shll.regs[i] = newReg()
	}
	return shll, nil
}

func (shll *SlidingHyperLogLog) valAndPos(value []byte) (uint8, uint64) {
	val := metro.Hash64(value, 32)
	k := 64 - shll.p
	j := val >> uint(k)
	R := rho(val<<shll.p, 6)
	return R, j
}

/*
Insert a value with a timestamp to the SlidingHyperLogLog.
*/
func (shll *SlidingHyperLogLog) Insert(timestamp uint64, value []byte) {
	R, j := shll.valAndPos(value)
	shll.regs[j].insert(tR{timestamp, R})
	if timestamp > shll.latestTimestamp {
		shll.latestTimestamp = timestamp
	}
}

/*
Estimate returns the estimated cardinality since a given timestamp
*/
func (shll *SlidingHyperLogLog) Estimate(timestamp uint64) uint64 {
	m := float64(shll.m)
	sum := regSumSince(shll.regs, timestamp)
	ez := zerosSince(shll.regs, timestamp)
	beta := beta(ez)
	return uint64(shll.alpha * m * (m - ez) / (beta + sum))
}
