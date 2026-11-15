package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"time"

	"dbut.dev/bolt"
	_ "dbut.dev/bolt/all"
)

const samples = 101

var client = &http.Client{Timeout: 15 * time.Second}

var taskRE = regexp.MustCompile(`var task = '([^']+)'`)

type puzzle struct {
	Family   string  `json:"family"`
	Variant  string  `json:"variant"`
	Rows     int     `json:"rows"`
	Cols     int     `json:"cols"`
	Clues    []int   `json:"clues"`
	Solution string  `json:"solution"`
	Order    []int   `json:"order"`
	Micros   float64 `json:"micros"`
}

func fetch(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "bolt-trace (+https://github.com/dbut2/bolt)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	m := taskRE.FindSubmatch(body)
	if m == nil {
		return "", fmt.Errorf("%s: no task on page", url)
	}
	return string(m[1]), nil
}

func solve(f bolt.Family, variant, task string) (puzzle, error) {
	var rows, cols int
	if _, err := fmt.Sscanf(variant, "%dx%d", &cols, &rows); err != nil {
		return puzzle{}, fmt.Errorf("variant %q: %w", variant, err)
	}
	clues := f.Parse(task)
	if len(clues) != rows*cols {
		return puzzle{}, fmt.Errorf("%s %s: parsed %d cells, want %d", f.Name, variant, len(clues), rows*cols)
	}

	warm := f.New()
	warm.Setup(clues, rows, cols)
	solution := warm.Solve()

	times := make([]time.Duration, samples)
	for i := range times {
		start := time.Now()
		s := f.New()
		s.Setup(clues, rows, cols)
		s.Solve()
		times[i] = time.Since(start)
	}
	slices.Sort(times)
	micros := float64(times[samples/2].Nanoseconds()) / 1000

	p := puzzle{f.Name, variant, rows, cols, clues, solution, []int{}, micros}
	traced := f.New()
	traced.Setup(clues, rows, cols)
	if t, ok := traced.(bolt.Tracer); ok {
		t.Trace(func(s bolt.Step) { p.Order = append(p.Order, s.Cell) })
		traced.Solve()
	}
	return p, nil
}

func run(n int, variant string, pause time.Duration) ([]puzzle, error) {
	puzzles := []puzzle{}
	for _, f := range bolt.Families() {
		path, ok := f.Variants[variant]
		if !ok {
			return nil, fmt.Errorf("%s: no variant %q", f.Name, variant)
		}

		seen := map[string]bool{}
		var failed error
		for attempts := 0; len(seen) < n; attempts++ {
			if attempts == 3*n {
				return nil, errors.Join(fmt.Errorf("%s: only %d distinct puzzles in %d fetches", f.Name, len(seen), attempts), failed)
			}
			if attempts > 0 {
				time.Sleep(pause)
			}
			task, err := fetch(f.Site + path)
			if err != nil {
				failed = err
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if seen[task] {
				continue
			}
			seen[task] = true

			p, err := solve(f, variant, task)
			if err != nil {
				return nil, err
			}
			puzzles = append(puzzles, p)
		}
		time.Sleep(pause)
	}
	return puzzles, nil
}

func main() {
	n := flag.Int("n", 10, "puzzles to fetch per family")
	variant := flag.String("variant", "10x10-easy", "variant to fetch for every family")
	pause := flag.Duration("pause", time.Second, "pause between requests")
	flag.Parse()

	puzzles, err := run(*n, *variant, *pause)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(puzzles); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
