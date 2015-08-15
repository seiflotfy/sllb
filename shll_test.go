package shll

import (
	"strconv"
	"testing"
)

func TestAdd(t *testing.T) {
	shll, err := NewSlidingHyperLogLog(0.05, 100)
	if err != nil {
		t.Error("Expected no error on NewSlidingHyperLogLog, got", err)
	}
	for i := 0; i < 10; i++ {
		shll.Add(uint32(i), []byte(strconv.Itoa(i)))
	}
}
