package virtual

type SystemModel struct {
	loadQueue      chan sourceLoad
	swingloadQueue chan Load
}

type sourceLoad struct {
	id   int
	load Load
}

type Load struct {
	Real     float64
	Reactive float64
}

//func NewVirtualSystemModel() SystemModel

func (s *SystemModel) ReportLoad(id int, load Load) {
	s.loadQueue <- sourceLoad{id, load}
}
