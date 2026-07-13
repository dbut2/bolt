package binairo

import (
	"iter"
	"strings"

	"github.com/dbut2/advent-of-code/pkg/space"

	"dbut.dev/bolt"
)

func init() {
	bolt.Register(bolt.Family{
		Name: "binairo",
		Site: "https://www.puzzle-binairo.com",
		Variants: map[string]string{
			"6x6-easy":   "/binairo-6x6-easy/",
			"8x8-easy":   "/binairo-8x8-easy/",
			"10x10-easy": "/binairo-10x10-easy/",
			"14x14-easy": "/binairo-14x14-easy/",
			"20x20-easy": "/binairo-20x20-easy/",
		},
		New:   func() bolt.Solver { return new(Board) },
		Parse: func(task string) []int { return bolt.ParseTask(task, -1) },
	})
}

type Board struct {
	grid space.Grid[int]
}

func (b *Board) Setup(cells []int, rows, cols int) {
	g := space.NewGrid[int](rows, cols)
	for i, cell := range cells {
		g.Set(space.Cell{i / cols, i % cols}, cell)
	}
	b.grid = g
}

const (
	unknown = iota - 1
	white
	black
)

func (b *Board) Solve() string {
	g := b.grid

	changed := true
	for changed {
		changed = false

		for cell, v := range cells(g) {
			if v == nil || *v != unknown {
				continue
			}

			uc, dc, lc, rc := cell.Move(space.Up), cell.Move(space.Down), cell.Move(space.Left), cell.Move(space.Right)
			uuc, ddc, llc, rrc := uc.Move(space.Up), dc.Move(space.Down), lc.Move(space.Left), rc.Move(space.Right)

			u, d, l, r := g.Get(uc), g.Get(dc), g.Get(lc), g.Get(rc)
			uu, dd, ll, rr := g.Get(uuc), g.Get(ddc), g.Get(llc), g.Get(rrc)

			if u != nil && d != nil && *u != unknown && *u == *d {
				*v = 1 - *u
				changed = true
				continue
			}

			if l != nil && r != nil && *l != unknown && *l == *r {
				*v = 1 - *l
				changed = true
				continue
			}

			if u != nil && uu != nil && *u != unknown && *u == *uu {
				*v = 1 - *u
				changed = true
				continue
			}

			if d != nil && dd != nil && *d != unknown && *d == *dd {
				*v = 1 - *d
				changed = true
				continue
			}

			if l != nil && ll != nil && *l != unknown && *l == *ll {
				*v = 1 - *l
				changed = true
				continue
			}

			if r != nil && rr != nil && *r != unknown && *r == *rr {
				*v = 1 - *r
				changed = true
				continue
			}
		}

		for i := range g {
			size := len(g)
			blacks, whites := 0, 0
			blanks := []space.Cell{}

			for j := range g[i] {
				v := g[i][j]
				switch v {
				case black:
					blacks++
				case white:
					whites++
				case unknown:
					blanks = append(blanks, space.Cell{i, j})
				}
			}

			if len(blanks) == 0 {
				continue
			}

			if blacks == size/2 {
				for _, cell := range blanks {
					g.Set(cell, white)
				}
				changed = true
				continue
			}

			if whites == size/2 {
				for _, cell := range blanks {
					g.Set(cell, black)
				}
				changed = true
				continue
			}
		}

		for j := range g[0] {
			size := len(g)
			blacks, whites := 0, 0
			blanks := []space.Cell{}

			for i := range g {
				v := g[i][j]
				switch v {
				case black:
					blacks++
				case white:
					whites++
				case unknown:
					blanks = append(blanks, space.Cell{i, j})
				}
			}

			if len(blanks) == 0 {
				continue
			}

			if blacks == size/2 {
				for _, cell := range blanks {
					g.Set(cell, white)
				}
				changed = true
				continue
			}

			if whites == size/2 {
				for _, cell := range blanks {
					g.Set(cell, black)
				}
				changed = true
				continue
			}
		}
	}

	out := strings.Builder{}
	out.Grow(len(b.grid) * len(b.grid))
	for _, line := range b.grid {
		for _, cell := range line {
			out.WriteByte(byte(cell + '0'))
		}
	}
	return out.String()
}

func cells[T any](g space.Grid[T]) iter.Seq2[space.Cell, *T] {
	return func(yield func(space.Cell, *T) bool) {
		for i := range g {
			for j := range g[i] {
				if !yield(space.Cell{i, j}, &g[i][j]) {
					return
				}
			}
		}
	}
}
