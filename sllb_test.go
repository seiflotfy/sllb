package sllb

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func sumFromIndex(counts []uint64, index uint64) uint64 {
	var count uint64
	for i := int(index); i < len(counts); i++ {
		count += counts[i]
	}
	return count
}

func TestInsertEstimate(t *testing.T) {
	sllb, err := New(0.008)
	if err != nil {
		t.Error("Expected no error on NewSlidingHyperLogLog, got", err)
	}

	counts := make([]uint64, 100)
	for i := 0; i < len(counts); i++ {
		for j := 0; j <= rand.Intn(100000); j++ {
			e := fmt.Sprintf("e-%d-%d", i, j)
			sllb.Insert(uint64(i), []byte(e))
			counts[i]++
		}
	}

	for i := uint64(0); i <= uint64(len(counts)); i++ {
		est := sllb.Estimate(i)
		exp := sumFromIndex(counts, i)
		offset := uint64(math.Abs(5 * float64(exp) / 100))
		if est < exp-offset || est > exp+offset {
			t.Errorf("%d Expected error <= 5.0%% for %d, got %d", i, exp, est)
		}
	}
}
