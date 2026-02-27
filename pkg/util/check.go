package util

import (
	"strings"
	"unicode"
)

func Exist[T ~int | ~int32 | ~uint32 | ~int64 | ~string](arr []T, n T) bool {
	for _, i := range arr {
		if i == n {
			return true
		}
	}
	return false
}

func IsAlphanumeric(input string) bool {
	for _, r := range input {
		if r == ':' || r == '"' || r == '\'' || unicode.IsUpper(r) {
			return false
		}
		if !unicode.Is(unicode.Han, r) && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' {
			if strings.ContainsAny(string(r), "~#@$%^&*()<>,.{}[]、|/？?;'!！=+") {
				return false
			}
		}
	}
	return true
}
