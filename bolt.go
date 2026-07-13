package bolt

import (
	"slices"
)

type Solver interface {
	Setup(cells []int, rows, cols int)
	Solve() string
}

type Family struct {
	Name     string
	Site     string
	Variants map[string]string
	New      func() Solver
	Parse    func(task string) []int
}

var families []Family

func Register(f Family) {
	families = append(families, f)
}

func Families() []Family {
	return slices.Clone(families)
}
