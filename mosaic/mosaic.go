package mosaic

import (
	"iter"
	"strings"

	"github.com/dbut2/advent-of-code/pkg/lists"
	"github.com/dbut2/advent-of-code/pkg/space"

	"dbut.dev/bolt"
)

func init() {
	bolt.Register(bolt.Family{
		Name: "mosaic",
		Site: "https://www.puzzle-minesweeper.com",
		Variants: map[string]string{
			"5x5-easy":   "/mosaic-5x5-easy/",
			"7x7-easy":   "/mosaic-7x7-easy/",
			"10x10-easy": "/mosaic-10x10-easy/",
			"15x15-easy": "/mosaic-15x15-easy/",
			"20x20-easy": "/mosaic-20x20-easy/",
		},
		New:   func() bolt.Solver { return new(Board) },
		Parse: func(task string) []int { return bolt.ParseTask(task, noClue) },
	})
}

const noClue = -1

const (
	unknown = iota
	on
	off
)

type Board struct {
	clues space.Grid[int]
	state space.Grid[int]
	trace func(bolt.Step)
}

func (b *Board) Trace(f func(bolt.Step)) {
	b.trace = f
}

func (b *Board) Setup(cells []int, rows, cols int) {
	clues := space.NewGrid[int](rows, cols)
	for i, cell := range cells {
		clues.Set(space.Cell{i / cols, i % cols}, cell)
	}
	b.clues = clues
	b.state = space.NewGrid[int](rows, cols)
}

func (b *Board) Solve() string {
	b.solve()

	out := strings.Builder{}
	out.Grow(len(b.state) * len(b.state[0]))
	for _, line := range b.state {
		for _, cell := range line {
			switch cell {
			case on:
				out.WriteByte('y')
			default:
				out.WriteByte('n')
			}
		}
	}
	return out.String()
}

func (b *Board) solve() {
	q := lists.NewQueue[space.Cell]()
	for i := range b.clues {
		for j := range b.clues[i] {
			q.Push(space.Cell{i, j})
		}
	}

	for cell := range q.Seq {
		clue := *b.clues.Get(cell)
		if clue == noClue {
			continue
		}

		uncs := lists.Queue[space.Cell]{}
		unc, onc := 0, 0

		switch *b.state.Get(cell) {
		case on:
			onc++
		case unknown:
			unc++
			uncs.Push(cell)
		}

		for n, v := range surrounding(b.state, cell) {
			switch *v {
			case on:
				onc++
			case unknown:
				unc++
				uncs.Push(n)
			}
		}

		if unc == 0 {
			continue
		}

		set, value := off, byte('n')
		switch clue {
		case onc:
		case onc + unc:
			set, value = on, 'y'
		default:
			continue
		}

		for next := range uncs.Seq {
			b.state.Set(next, set)
			if b.trace != nil {
				b.trace(bolt.Step{Cell: next[0]*len(b.state[0]) + next[1], Value: value})
			}
			q.Push(next)
			for n := range surrounding(b.state, next) {
				q.Push(n)
			}
		}
	}
}

func surrounding(g space.Grid[int], c space.Cell) iter.Seq2[space.Cell, *int] {
	return func(yield func(space.Cell, *int) bool) {
		for i := max(c[0]-1, 0); i <= min(c[0]+1, len(g)-1); i++ {
			for j := max(c[1]-1, 0); j <= min(c[1]+1, len(g[i])-1); j++ {
				if i == c[0] && j == c[1] {
					continue
				}
				if !yield(space.Cell{i, j}, &g[i][j]) {
					return
				}
			}
		}
	}
}
