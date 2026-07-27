package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type variant struct {
	label    string
	hallsize int
}

type site struct {
	family   string
	base     string
	variants []variant
}

var sites = []site{
	{"binairo", "https://www.puzzle-binairo.com", []variant{
		{"6x6 easy", 0},
		{"8x8 easy", 2},
		{"10x10 easy", 4},
		{"14x14 easy", 6},
		{"20x20 easy", 8},
	}},
	{"dominosa", "https://www.puzzle-dominosa.com", []variant{
		{"3x3", 0},
		{"4x4", 6},
		{"5x5", 7},
		{"6x6", 1},
		{"7x7", 8},
		{"8x8", 9},
		{"9x9", 2},
	}},
	{"hitori", "https://www.puzzle-hitori.com", []variant{
		{"5x5 easy", 0},
		{"10x10 easy", 3},
		{"15x15 easy", 6},
		{"20x20 easy", 9},
	}},
	{"minesweeper", "https://www.puzzle-minesweeper.com", []variant{
		{"5x5 easy", 0},
		{"7x7 easy", 2},
		{"10x10 easy", 4},
		{"15x15 easy", 6},
		{"20x20 easy", 8},
	}},
	{"mosaic", "https://www.puzzle-minesweeper.com", []variant{
		{"5x5 easy", 13},
		{"7x7 easy", 15},
		{"10x10 easy", 17},
		{"15x15 easy", 19},
		{"20x20 easy", 21},
	}},
	{"nurikabe", "https://www.puzzle-nurikabe.com", []variant{
		{"5x5 easy", 0},
		{"7x7 easy", 1},
		{"10x10 easy", 2},
	}},
	{"shikaku", "https://www.puzzle-shikaku.com", []variant{
		{"5x5", 0},
		{"7x7", 1},
		{"10x10", 2},
		{"15x15", 3},
		{"20x20", 4},
		{"25x25", 5},
	}},
	{"slitherlink", "https://www.puzzle-loop.com", []variant{
		{"5x5 normal", 0},
		{"7x7 normal", 10},
		{"10x10 normal", 1},
		{"15x15 normal", 2},
		{"20x20 normal", 3},
		{"25x30 normal", 8},
	}},
	{"sudoku", "https://www.puzzle-sudoku.com", []variant{
		{"3x3 easy", 1},
		{"3x3 intermediate", 2},
		{"3x3 advanced", 3},
		{"3x3 extreme", 4},
		{"3x3 evil", 5},
	}},
}

const heading = "Robots / programmatic solvers"

var rowRE = regexp.MustCompile(`(?s)<tr><td[^>]*>(\d+)\.</td><td class="nick"[^>]*>(?:<b>)?([^<]+).*?<td align="right">(?:<b>)?(\d\d:\d\d\.\d+)`)

type entry struct {
	rank int
	name string
	time string
}

func robots(url string) ([]entry, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "bolt-rankings (+https://github.com/dbut2/bolt)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	start := strings.Index(html, heading)
	if start == -1 {
		return nil, fmt.Errorf("no robots section on page")
	}
	section := html[start:]
	if end := strings.Index(section, "</table>"); end != -1 {
		section = section[:end]
	}

	var entries []entry
	for _, m := range rowRE.FindAllStringSubmatch(section, -1) {
		rank, _ := strconv.Atoi(m[1])
		entries = append(entries, entry{rank, strings.TrimSpace(m[2]), m[3]})
	}
	return entries, nil
}

var medals = map[int]string{1: "🥇", 2: "🥈", 3: "🥉"}

func main() {
	readme := flag.String("readme", "", "README file to update between the rankings markers")
	flag.Parse()

	bot := os.Getenv("BOT_NAME")
	if bot == "" {
		bot = "dbut2"
	}

	var rows [][4]string
	podiums, boards := 0, 0
	for _, s := range sites {
		for _, v := range s.variants {
			boards++
			url := fmt.Sprintf("%s/hall.php?hallsize=%d", s.base, v.hallsize)
			puzzle := fmt.Sprintf("[%s %s](%s)", s.family, v.label, url)
			entries, err := robots(url)
			if err != nil {
				rows = append(rows, [4]string{puzzle, "⚠️ error", err.Error(), ""})
				continue
			}

			leader := "—"
			if len(entries) > 0 {
				leader = fmt.Sprintf("%s (%s)", entries[0].name, entries[0].time)
			}

			row := [4]string{puzzle, "not listed", "—", leader}
			for _, e := range entries {
				if e.name == bot {
					if e.rank <= 3 {
						podiums++
					}
					row[1] = strings.TrimSpace(fmt.Sprintf("%s %d", medals[e.rank], e.rank))
					row[2] = e.time
					break
				}
			}
			rows = append(rows, row)
			time.Sleep(500 * time.Millisecond)
		}
	}

	report := &strings.Builder{}
	fmt.Fprintf(report, "## Bot rankings — `%s`\n\n", bot)
	fmt.Fprintf(report, "Top 3 on **%d of %d** tracked boards.\n\n", podiums, boards)
	report.WriteString("| Puzzle | Rank | Time | Leader |\n|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(report, "| %s | %s | %s | %s |\n", r[0], r[1], r[2], r[3])
	}

	if summary := os.Getenv("GITHUB_STEP_SUMMARY"); summary != "" {
		f, err := os.OpenFile(summary, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		if _, err := f.WriteString(report.String()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if *readme != "" {
		if err := updateReadme(*readme, report.String()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Print(report.String())
}

const (
	markerBegin = "<!-- rankings:begin -->"
	markerEnd   = "<!-- rankings:end -->"
)

func updateReadme(path, report string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(src)
	begin := strings.Index(s, markerBegin)
	end := strings.Index(s, markerEnd)
	if begin == -1 || end == -1 || end < begin {
		return fmt.Errorf("%s: missing %s / %s markers", path, markerBegin, markerEnd)
	}
	stamped := fmt.Sprintf("\n%s\n_Updated %s._\n", strings.TrimSpace(report), time.Now().UTC().Format("2 Jan 2006 15:04 UTC"))
	out := s[:begin+len(markerBegin)] + stamped + s[end:]
	return os.WriteFile(path, []byte(out), 0o644)
}
