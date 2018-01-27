package manifest

import (
	"strings"
	"unicode"
)

func Heredoc(raw string) (res string) {
	minIndentSize := int(^uint(0) >> 1) // Max value of type int
	lines := strings.Split(raw, "\n")
	if len(lines) == 1 {
		return lines[0]
	}
	if strings.TrimSpace(string(raw[0])) == "" {
		lines = lines[0:]
	}

	// 1.
	for i, line := range lines {
		indentSize := 0
		for _, r := range []rune(line) {
			if unicode.IsSpace(r) {
				indentSize += 1
			} else {
				break
			}
		}

		if len(line) == indentSize {
			if i == len(lines)-1 && indentSize < minIndentSize {
				lines[i] = ""
			}
		} else if indentSize < minIndentSize {
			minIndentSize = indentSize
		}
	}

	// 2.
	for i, line := range lines {
		if len(lines[i]) >= minIndentSize {
			lines[i] = line[minIndentSize:]
		}
	}

	return strings.Join(lines, "\n")
}
