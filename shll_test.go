package shll

import (
	"fmt"
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
	fmt.Println(shll.GetCount(1000000, 0))
}
