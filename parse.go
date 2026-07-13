package bolt

import "unicode"

func ParseTask(task string, empty int) []int {
	var cells []int
	for _, char := range task {
		switch {
		case unicode.IsDigit(char):
			cells = append(cells, int(char-'0'))
		case unicode.IsLetter(char):
			for range char - 'a' + 1 {
				cells = append(cells, empty)
			}
		}
	}
	return cells
}
