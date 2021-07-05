package color

import "strings"

const (
	reset   = "\u001b[0m"
	black   = "\u001b[30m"
	red     = "\u001b[31m"
	green   = "\u001b[32m"
	yellow  = "\u001b[33m"
	blue    = "\u001b[34m"
	magenta = "\u001b[35m"
	cyan    = "\u001b[36m"
	white   = "\u001b[37m"
)

// Parse will convert color codes into ANSI colors.
func Parse(input string) string {
	str := input
	str = strings.ReplaceAll(str, "{x", reset)
	str = strings.ReplaceAll(str, "{k", black)
	str = strings.ReplaceAll(str, "{r", red)
	str = strings.ReplaceAll(str, "{g", green)
	str = strings.ReplaceAll(str, "{y", yellow)
	str = strings.ReplaceAll(str, "{b", blue)
	str = strings.ReplaceAll(str, "{m", magenta)
	str = strings.ReplaceAll(str, "{c", cyan)
	str = strings.ReplaceAll(str, "{w", white)
	return str
}

// Strip will remove color codes from a string.
func Strip(input string) string {
	str := input
	str = strings.ReplaceAll(str, "{x", "")
	str = strings.ReplaceAll(str, "{k", "")
	str = strings.ReplaceAll(str, "{r", "")
	str = strings.ReplaceAll(str, "{g", "")
	str = strings.ReplaceAll(str, "{y", "")
	str = strings.ReplaceAll(str, "{b", "")
	str = strings.ReplaceAll(str, "{m", "")
	str = strings.ReplaceAll(str, "{c", "")
	str = strings.ReplaceAll(str, "{w", "")
	return str
}
