package shll

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestAdd(t *testing.T) {
	shll, err := NewSlidingHyperLogLog(0.005, 1000000, 100)
	if err != nil {
		t.Error("Expected no error on NewSlidingHyperLogLog, got", err)
	}
	for i := 0; i < 1000000; i++ {
		shll.Add(uint32(i+1), []byte("test-"+strconv.Itoa(rand.Int())))
	}

	count, _ := shll.GetCount(1000000, 0)
	res := math.Abs(100 * (1 - float64(count)/1000000))
	if res > 3 {
		t.Errorf("Expected error <= 3.0%%, got %f", res)
	}

	count, _ = shll.GetCount(1000000, 100)
	res = math.Abs(100 * (1 - float64(count)/100))
	if res > 3 {
		t.Errorf("Expected error <= 3.0%%, got %f", res)
	}

	count, _ = shll.GetCount(0, 100)
	if count != 0 {
		t.Errorf("Expected error <= 0.0%%, got %f", res)
	}

	count, _ = shll.GetCount(500000, 1000)
	res = math.Abs(100 * (1 - float64(count)/100))
	if res > 19 {
		t.Errorf("Expected error <= 0.19%%, got %f", res)
	}
}
