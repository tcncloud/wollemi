package please

import (
	"bufio"
)

type Ctl interface {
	QueryDeps(...string) (*bufio.Reader, error)
	Graph() (*Graph, error)
	Build(...string) error
}
