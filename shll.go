package shll

import (
	"container/heap"
	"errors"
	"hash"
	"hash/fnv"
	"math"
)

/*
SlidingHyperLogLog ...
*/
type SlidingHyperLogLog struct {
	window uint32
	alpha  float64
	p      uint
	m      uint
	lpfm   [][]tR
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
NewSlidingHyperLogLog ...
*/
func NewSlidingHyperLogLog(errRate float64, window uint32) (*SlidingHyperLogLog, error) {
	if !(0 < errRate && errRate < 1) {
		return nil, errors.New("errRate must be between 0 and 1")
	}
	shll := &SlidingHyperLogLog{}
	shll.p = uint(math.Ceil(math.Log2(math.Pow((1.04 / errRate), 2))))
	shll.m = 1 << shll.p
	shll.lpfm = make([][]tR, shll.m, shll.m)
	shll.alpha = getAlpha(shll.m)
	shll.window = window
	shll.hasher = fnv.New64a()

	return shll, nil
}

/*
Add ...
*/
func (shll *SlidingHyperLogLog) Add(timestamp uint32, value []byte) {
	shll.hasher.Write(value)
	val := shll.hasher.Sum64()
	shll.hasher.Reset()
	k := 64 - shll.p
	j := val >> uint(k)
	R := getRho(val<<shll.p, 6)

	Rmax := uint8(0)
	var tmp []tR
	tmax := uint32(0)
	h := tRHeap(shll.lpfm[j])
	heap.Init(&h)
	heap.Push(&h, tR{timestamp, R})

	tmp2 := make([]tR, h.Len())
	for h.Len() > 0 {
		item := heap.Pop(&h).(tR)
		tmp2[h.Len()] = item
	}

	for i, value := range tmp2 {
		t := value.t
		R := value.R
		if tmax == 0 {
			tmax = t
		}
		if t < (tmax - shll.window) {
			break
		}
		if R > Rmax {
			tmp[i] = value
			Rmax = R
		}
	}

	shll.lpfm[j] = tmp
}

/*
GetCount ...
*/
func (shll *SlidingHyperLogLog) GetCount(timestamp uint32, window uint32) (uint32, error) {
	/*
		"""
		Returns the estimate of the cardinality at 'timestamp' using 'window'
		"""
		if window is None:
		    window = self.window

		if not 0 < window <= self.window:
		    raise ValueError('0 < window <= W')

		def max_r(l):
		    return max(l) if l else 0

		M = [max_r([R for ts, R in lpfm if ts >= (timestamp - window)]) if lpfm else 0 for lpfm in self.LPFM]

		#count number or registers equal to 0
		V = M.count(0)
		if V > 0:
		    H = self.m * math.log(self.m / float(V))
		    return H if H <= get_treshold(self.p) else self._Ep(M)
		else:
		    return self._Ep(M)
	*/
	if window == 0 {
		window = shll.window
	}
	if !(0 < window && window <= shll.window) {
		return 0, errors.New("0 < window <= W")
	}

	var maxR = func(l []uint) uint {
		if len(l) == 0 {
			return 0
		}
		temp := uint(0)
		for _, value := range l {
			if value > temp {
				temp = value
			}
		}
		return temp
	}

	//M = [max_r([R for ts, R in lpfm if ts >= (timestamp - window)]) if lpfm else 0 for lpfm in self.LPFM]
	M := make([]uint8, len(shll.lpfm), len(shll.lpfm))
	for i, lfpm := range shll.lpfm {
		if len(lfpm) == 0 {
			M[i] = 0
			continue
		}
		var tempM []uint8
		M[i] = maxR()
	}
	return 0, nil
}
