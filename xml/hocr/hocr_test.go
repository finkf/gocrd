package hocr

func withOpenPcGts(path string, f func(p *PcGts)) {
	p, err := Open(path)
	if err != nil {
		panic(err)
	}
	f(p)
}
