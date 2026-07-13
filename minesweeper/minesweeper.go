package minesweeper

import (
	"strings"

	"github.com/dbut2/advent-of-code/pkg/lists"
	"github.com/dbut2/advent-of-code/pkg/space"

	"dbut.dev/bolt"
)

func init() {
	bolt.Register(bolt.Family{
		Name: "minesweeper",
		Site: "https://www.puzzle-minesweeper.com",
		Variants: map[string]string{
			"5x5-easy":   "/minesweeper-5x5-easy/",
			"7x7-easy":   "/minesweeper-7x7-easy/",
			"10x10-easy": "/minesweeper-10x10-easy/",
			"15x15-easy": "/minesweeper-15x15-easy/",
			"20x20-easy": "/minesweeper-20x20-easy/",
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

func (b *Board) Solve() string {
	q := lists.NewStack[space.Cell]()
	for cell := range b.grid.Cells() {
		q.Push(cell)
	}

	for cell := range q.Seq {
		b.step(cell, &q)
	}

	out := strings.Builder{}
	out.Grow(len(b.grid) * len(b.grid))
	for _, line := range b.grid {
		for _, cell := range line {
			switch cell {
			case flag:
				out.WriteByte('y')
			default:
				out.WriteByte('n')
			}
		}
	}
	return out.String()
}

const (
	unknown = -1 - iota
	flag
	empty
)

func (b *Board) step(c space.Cell, q *lists.Stack[space.Cell]) {
	v := b.grid.Get(c)
	if v == nil || *v < 0 {
		return
	}

	var flags, _, unknowns int

	for _, nv := range b.grid.Surrounding(c) {
		if nv == nil {
			continue
		}

		switch *nv {
		case unknown:
			unknowns++
		case flag:
			flags++
		}
	}

	if unknowns == 0 {
		return
	}

	// case remaining flags
	if *v == unknowns+flags {
		for nc, nv := range b.grid.Surrounding(c) {
			if nv == nil || *nv != unknown {
				continue
			}
			b.grid.Set(nc, flag)
			for nnc, nnv := range b.grid.Surrounding(nc) {
				if nnv == nil || *nnv < 0 {
					continue
				}
				q.Push(nnc)
			}
		}
		return
	}

	// case remaining empty
	if *v == flags {
		for nc, nv := range b.grid.Surrounding(c) {
			if nv == nil || *nv != unknown {
				continue
			}
			b.grid.Set(nc, empty)
			for nnc, nnv := range b.grid.Surrounding(nc) {
				if nnv == nil || *nnv < 0 {
					continue
				}
				q.Push(nnc)
			}
		}
		return
	}
}
