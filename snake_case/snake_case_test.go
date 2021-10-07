package snake_case

import (
	"regexp"
	"strings"
	"testing"
	"unicode"
)

var cases = []struct {
	args string
	want string
}{
	{"", ""},
	{"camelCase", "camel_case"},
	{"PascalCase", "pascal_case"},
	{"snake_case", "snake_case"},
	{"Pascal_Snake", "pascal_snake"},
	{"SCREAMING_SNAKE", "screaming_snake"},
	{"kebab-case", "kebab_case"},
	{"Pascal-Kebab", "pascal_kebab"},
	{"SCREAMING-KEBAB", "screaming_kebab"},
	{"A", "a"},
	{"AA", "aa"},
	{"AAA", "aaa"},
	{"AAAA", "aaaa"},
	{"AaAa", "aa_aa"},
	{"HTTPRequest", "http_request"},
	{"BatteryLifeValue", "battery_life_value"},
	{"Id0Value", "id0_value"},
	{"ID0Value", "id0_value"},
}

func TestToSnakeCase(t *testing.T) {
	for _, tt := range cases {
		t.Run("ToSnakeCase: "+tt.args, func(t *testing.T) {
			if got := ToSnakeCase(tt.args); got != tt.want {
				t.Errorf("ToSnakeCase(%#q) = %#q, want %#q", tt.args, got, tt.want)
			}
		})
	}
}

func BenchmarkCaseByCase(b *testing.B) {
	for _, item := range cases {
		arg := item.args
		b.Run(arg, func(b *testing.B) {
			b.Run("ToSnakeCaseRegex", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ToSnakeCaseRegex(arg)
				}
			})
			b.Run("ToSnakeCaseByJensSkipr", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ToSnakeCaseByJensSkipr(arg)
				}
			})
			b.Run("ToSnakeCase", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ToSnakeCase(arg)
				}
			})
		})
	}
}

func BenchmarkOneCase(b *testing.B) {
	b.Run("ToSnakeCaseRegex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToSnakeCaseRegex("BatteryLifeValue")
		}
	})
	b.Run("ToSnakeCaseByJensSkipr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToSnakeCaseByJensSkipr("BatteryLifeValue")
		}
	})
	b.Run("ToSnakeCase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToSnakeCase("BatteryLifeValue")
		}
	})
}

func BenchmarkAllInOne(b *testing.B) {
	b.Run("ToSnakeCaseRegex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, item := range cases {
				arg := item.args
				ToSnakeCaseRegex(arg)
			}
		}
	})
	b.Run("ToSnakeCaseByJensSkipr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, item := range cases {
				arg := item.args
				ToSnakeCaseByJensSkipr(arg)
			}
		}
	})
	b.Run("ToSnakeCase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, item := range cases {
				arg := item.args
				ToSnakeCase(arg)
			}
		}
	})
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCaseRegex(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ToSnakeCaseByJensSkipr(s string) string {
	var res = make([]rune, 0, len(s))
	var p = '_'
	for i, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			res = append(res, '_')
		} else if unicode.IsUpper(r) && i > 0 {
			if unicode.IsLetter(p) && !unicode.IsUpper(p) || unicode.IsDigit(p) {
				res = append(res, '_', unicode.ToLower(r))
			} else {
				res = append(res, unicode.ToLower(r))
			}
		} else {
			res = append(res, unicode.ToLower(r))
		}

		p = r
	}
	return string(res)
}
