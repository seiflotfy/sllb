package shll

import (
	"container/heap"
	"errors"
	"hash"
	"hash/fnv"
	"math"
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
	window uint32
	alpha  float64
	p      uint
	m      uint
	lpfm   []*tRHeap
	n      uint
	hasher hash.Hash64
}

type tR struct {
	t uint32 //timestamp
	R uint8  //trailing 0s
}

type tRHeap []tR

func (h tRHeap) Len() int           { return len(h) }
func (h tRHeap) Less(i, j int) bool { return h[i].t < h[j].t }
func (h tRHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *tRHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(tR))
}

func (h *tRHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func getRho(val uint64, max uint32) uint8 {
	r := uint32(1)
	for val&0x80000000 == 0 && r <= max {
		r++
		val <<= 1
	}
	return uint8(r)
}

func getAlpha(m uint) (result float64) {
	switch m {
	case 16:
		result = 0.673
	case 32:
		result = 0.697
	case 64:
		result = 0.709
	default:
		result = 0.7213 / (1.0 + 1.079/float64(m))
	}
	return result
}

/*
NewSlidingHyperLogLog return a new SlidingHyperLogLog.
errorRate = abs_err / cardinality
window = window to keep track of
nFlows = number of possible future maximum to keep track of per regiser
*/
func NewSlidingHyperLogLog(errRate float64, window uint32, nFlows uint32) (*SlidingHyperLogLog, error) {
	if !(0 < errRate && errRate < 1) {
		return nil, errors.New("errRate must be between 0 and 1")
	}
	shll := &SlidingHyperLogLog{}
	shll.p = uint(math.Ceil(math.Log2(math.Pow((1.04 / errRate), 2))))
	shll.m = 1 << shll.p
	shll.lpfm = make([]*tRHeap, shll.m, shll.m)
	for i := range shll.lpfm {
		shll.lpfm[i] = &tRHeap{}
		heap.Init(shll.lpfm[i])
	}
	shll.alpha = getAlpha(shll.m)
	shll.window = window
	shll.hasher = fnv.New64a()

	return shll, nil
}

func (shll *SlidingHyperLogLog) getPosAndValue(value []byte) (uint8, uint64) {
	shll.hasher.Write(value)
	val := shll.hasher.Sum64()
	shll.hasher.Reset()
	k := 64 - shll.p
	j := val >> uint(k)
	R := getRho(val<<shll.p, 6)
	return R, j
}

/*
Add a value with a timestamp to the SlidingHyperLogLog.
*/
func (shll *SlidingHyperLogLog) Add(timestamp uint32, value []byte) {
	R, j := shll.getPosAndValue(value)
	Rmax := uint8(0)
	tmax := int(0)
	heap.Push(shll.lpfm[j], tR{timestamp, R})

	tmp2 := make([]tR, shll.lpfm[j].Len(), shll.lpfm[j].Len())
	for shll.lpfm[j].Len() > 0 {
		item := heap.Pop(shll.lpfm[j]).(tR)
		tmp2[shll.lpfm[j].Len()] = item
	}

	for _, value := range tmp2 {
		t := value.t
		R := value.R

		if tmax == 0 {
			tmax = int(t)
		}
		if int(t) < (tmax - int(shll.window)) {
			break
		}

		if R > Rmax {
			Rmax = R
			heap.Push(shll.lpfm[j], value)
		}

		if uint(shll.lpfm[j].Len()) == shll.n {
			break
		}

	}

}

/*
GetCount returns the estimated cardinality within a 'window' to the past from a given 'timestamp'
*/
func (shll *SlidingHyperLogLog) GetCount(timestamp uint32, window uint32) (uint, error) {
	if window == 0 {
		window = shll.window
	}

	if !(0 < window && window <= shll.window) {
		return 0, errors.New("0 < window <= W")
	}

	var maxR = func(l []uint8) uint8 {
		temp := uint8(0)
		for _, value := range l {
			if value > temp {
				temp = value
			}
		}
		return temp
	}

	m := float64(shll.m)
	v := 0
	sum := 0.0
	M := make([]uint8, len(shll.lpfm), len(shll.lpfm))
	for i, lfpm := range shll.lpfm {
		if lfpm.Len() == 0 {
			M[i] = 0
			continue
		}
		Rs := make([]uint8, lfpm.Len(), lfpm.Len())
		tmp2 := make([]tR, lfpm.Len(), lfpm.Len())
		for lfpm.Len() > 0 {
			item := heap.Pop(lfpm).(tR)
			tmp2[lfpm.Len()] = item
		}

		for i, tr := range tmp2 {
			if int(tr.t) >= int(timestamp)-int(window) &&
				int(tr.t) <= int(timestamp) {
				Rs[i] = tr.R
			}
			heap.Push(lfpm, tr)
		}
		M[i] = maxR(Rs)
	}

	for _, value := range M {
		sum += 1.0 / math.Pow(2.0, float64(value))
		if value == 0 {
			v++
		}
	}

	estimate := shll.alpha * m * m / sum
	if estimate <= 5.0/2.0*m {
		// Small range correction
		if v > 0 {
			estimate = m * math.Log(m/float64(v))
		}
	} else if estimate > 1.0/30.0*exp32 {
		// Large range correction
		estimate = -exp32 * math.Log(1-estimate/exp32)
	}
	return uint(estimate), nil
}
