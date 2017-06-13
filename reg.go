package shll

type tR struct {
	t uint64 //timestamp
	R uint8  //trailing 0s
}

type reg struct {
	lfpm []tR
}

func newReg() *reg {
	return &reg{
		lfpm: make([]tR, 0),
	}
}

func (r *reg) insert(tr tR) {
	nlfpm := make([]tR, 0)
	for _, v := range r.lfpm {
		if v.t < tr.t && v.R < tr.R {
			continue
		}
		nlfpm = append(nlfpm, v)
	}
	r.lfpm = append(nlfpm, tr)
}

func (r *reg) get(timestamp uint64) uint8 {
	var val uint8
	for _, v := range r.lfpm {
		if v.t >= timestamp {
			if v.R > val {
				val = v.R
			}
		}
	}
	return val
}
