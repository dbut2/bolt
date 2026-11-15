package bolt_test

import (
	"testing"

	"dbut.dev/bolt"
	_ "dbut.dev/bolt/all"
)

var tasks = map[string]string{
	"mosaic":      "a3j4a5a655a2345a6c1a34a5b33a2b5435b10246a5c2b5a6666a3a4a57a64a3a52b543a2b02b32b",
	"minesweeper": "c3a3d1a2d31b334e2d22c11a2b4a2a1b2d2b2a2b5a32b33d34a2a3b1b3k",
	"binairo":     "a1a1d1a0b1a0d0l1c00a1a11b1g0h0d10b1b1b1a1j1c0d",
}

func BenchmarkSolve(b *testing.B) {
	for _, f := range bolt.Families() {
		cells := f.Parse(tasks[f.Name])
		b.Run(f.Name, func(b *testing.B) {
			for b.Loop() {
				s := f.New()
				s.Setup(cells, 10, 10)
				s.Solve()
			}
		})
	}
}

func TestTrace(t *testing.T) {
	for _, f := range bolt.Families() {
		t.Run(f.Name, func(t *testing.T) {
			cells := f.Parse(tasks[f.Name])
			s := f.New()
			s.Setup(cells, 10, 10)
			var steps []bolt.Step
			s.(bolt.Tracer).Trace(func(step bolt.Step) { steps = append(steps, step) })
			solution := s.Solve()

			if len(steps) == 0 {
				t.Fatal("no steps traced")
			}
			seen := map[int]bool{}
			for _, step := range steps {
				if seen[step.Cell] {
					t.Fatalf("cell %d settled twice", step.Cell)
				}
				seen[step.Cell] = true
				if solution[step.Cell] != step.Value {
					t.Fatalf("cell %d traced as %q, solved as %q", step.Cell, step.Value, solution[step.Cell])
				}
			}

			plain := f.New()
			plain.Setup(cells, 10, 10)
			if got := plain.Solve(); got != solution {
				t.Fatalf("traced solution %q differs from untraced %q", solution, got)
			}
		})
	}
}
