package azkaban

type Paginator struct {
	Length int
	Offset int
}

func NMostRecent(n int) Paginator {
	return Paginator{
		Length: n,
		Offset: 0,
	}
}

var TenMostRecent = Paginator{Length: 10, Offset: 0}
