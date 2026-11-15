package bolt

import (
	"slices"
)

type Solver interface {
	Setup(cells []int, rows, cols int)
	Solve() string
}

type Step struct {
	Cell  int
	Value byte
}

// Tracer is optionally implemented by a Solver. Solve reports each cell as it settles,
// with Cell as row*cols+col and Value as the byte that cell has in Solve's result.
type Tracer interface {
	Trace(func(Step))
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
