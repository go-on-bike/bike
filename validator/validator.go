package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type ValidationFunc func(interface{}) bool

// Expresiones regulares para validaciones comunes.
var (
    digitRX = regexp.MustCompile(`^[0-9]+$`)
    rutRX   = regexp.MustCompile(`^[0-9]{1,8}-[0-9Kk]$`)
    uuidRX  = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
	emailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func NotEmpty(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChar(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func MinChar(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func AllowedValues[T comparable](value T, allowedValues ...T) bool {
	return slices.Contains(allowedValues, value)
}

func StringUUID(value string) bool {
	match := uuidRX.MatchString(value)
	return match
}

func StringDigit(value string) bool {
	match := digitRX.MatchString(value)
	return match
}

func StringRut(value string) bool {
	match := rutRX.MatchString(value)
	return match
}

func StringEmail(value string) bool {
    match := emailRX.MatchString(value)
    return match
}
