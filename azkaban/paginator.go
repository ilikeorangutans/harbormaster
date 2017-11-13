package azkaban

type Paginator struct {
	Length int
	Offset int
}

var TenMostRecent = Paginator{Length: 10, Offset: 0}
