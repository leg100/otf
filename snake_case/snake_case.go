package snake_case

import (
	"strings"
	"unicode"
)

func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return strings.ToLower(s)
	}
	source := []rune(s)
	dist := strings.Builder{}
	dist.Grow(len(s) + len(s)/3) // avoid reallocation memory, 33% ~ 50% is recommended
	skipNext := false
	dist.WriteRune(unicode.ToLower(source[0]))
	last := source[1]
	for i := 2; i < len(source); i++ {
		cur := source[i]
		switch last {
		case '-', '_':
			dist.WriteRune('_')
			skipNext = true
			last = cur
			continue
		}
		if !unicode.IsLetter(last) {
			dist.WriteRune(last)
			dist.WriteRune('_')
			skipNext = true
			last = cur
			continue
		}
		if unicode.IsUpper(last) {
			if unicode.IsLower(cur) {
				if skipNext {
					skipNext = false
				} else {
					dist.WriteRune('_')
				}
			}
			dist.WriteRune(unicode.ToLower(last))
		} else {
			dist.WriteRune(last)
		}
		last = cur
	}

	dist.WriteRune(unicode.ToLower(source[len(source)-1]))
	return dist.String()
}
