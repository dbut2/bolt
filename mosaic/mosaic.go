package mosaic

import (
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
	for cell := range b.clues.Cells() {
		q.Push(cell)
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

		for n, v := range b.state.Surrounding(cell) {
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

		set := off
		switch clue {
		case onc:
		case onc + unc:
			set = on
		default:
			continue
		}

		for next := range uncs.Seq {
			b.state.Set(next, set)
			q.Push(next)
			for n := range b.state.Surrounding(next) {
				q.Push(n)
			}
		}
	}
}
