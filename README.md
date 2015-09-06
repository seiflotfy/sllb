# Sliding-HyperLogLog
[![GoDoc](https://godoc.org/github.com/seiflotfy/sliding-hyperloglog?status.svg)](https://godoc.org/github.com/seiflotfy/sliding-hyperloglog)

An implementation of an algorithm for estimating the number of active flows in a data stream is proposed. This algorithm adapts the HyperLogLog algorithm of Flajolet et. al to the data stream processing by adding a sliding window mechanism. It has the advantage to estimate at any time the number of flows seen over any duration bounded by the length of the sliding window.

The estimate is very accurate with a standard error of about 1.04/sqrt(m) (the same as the HyperLogLog algorithm). As the new algorithm answers more flexible queries, it needs an additional memory storage compared to HyerLogLog algorithm. It is proved that this additional memory is at most equal to 5m * ln(n/m) bytes, where n is the real number of flows in the sliding window. For instance, with an additional memory of only 35kB, a standard error of about 3% can be achieved for a data stream of several million flows. Theoretical results are validated on both real and synthetic traffic.

For details about the algorithm and citations please use this article for now:

["Sliding HyperLogLog: Estimating cardinality in a data stream" by Yousra Chabchoub & Georges Hébrail Kaminsky](https://hal.archives-ouvertes.fr/hal-00465313/file/sliding_HyperLogLog.pdf)

## Example Usage:
```go
shll, err := NewSlidingHyperLogLog(0.005, 1000000, 100) //created a Sliding HyperLogLog

shll.Add(1234567890, []byte("shll"))
shll.Add(1234567899, []byte("shll2"))

shll.GetCount(1234567890, 0)
//returns 2

```
