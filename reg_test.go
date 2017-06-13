package shll

import (
	"testing"
)

type testStruct struct {
	name  string
	tr    tR
	exptr []tR
}

var testData = []*testStruct{
	&testStruct{name: "A", tr: tR{t: 0, R: 5}, exptr: []tR{tR{t: 0, R: 5}}},
	&testStruct{name: "B", tr: tR{t: 1, R: 3}, exptr: []tR{tR{t: 0, R: 5}, tR{t: 1, R: 3}}},
	&testStruct{name: "C", tr: tR{t: 2, R: 4}, exptr: []tR{tR{t: 0, R: 5}, tR{t: 2, R: 4}}},
	&testStruct{name: "D", tr: tR{t: 3, R: 2}, exptr: []tR{tR{t: 0, R: 5}, tR{t: 2, R: 4}, tR{t: 3, R: 2}}},
	&testStruct{name: "E", tr: tR{t: 4, R: 1}, exptr: []tR{tR{t: 0, R: 5}, tR{t: 2, R: 4}, tR{t: 3, R: 2}, tR{t: 4, R: 1}}},
	&testStruct{name: "F", tr: tR{t: 5, R: 6}, exptr: []tR{tR{t: 5, R: 6}}},
}

func testEq(a, b []tR) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestInsert(t *testing.T) {
	r := newReg()
	for _, td := range testData {
		r.insert(td.tr)
		if !testEq(r.lfpm, td.exptr) {
			t.Errorf("expected %v, got %v", td.exptr, r.lfpm)
		}
	}
}
