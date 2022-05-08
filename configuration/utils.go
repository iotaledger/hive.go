package configuration

import (
	"runtime"
	"strings"
	"unicode"
)

func lowerCamelCase(str string) string {
	runes := []rune(str)
	runeCount := len(runes)

	if runeCount == 0 || unicode.IsLower(runes[0]) {
		return str
	}

	runes[0] = unicode.ToLower(runes[0])
	if runeCount == 1 || unicode.IsLower(runes[1]) {
		return string(runes)
	}

	for i := 1; i < runeCount; i++ {
		if i+1 < runeCount && unicode.IsLower(runes[i+1]) {
			break
		}

		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

func callerShortPackageName() string {
	pc, _, _, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	firstDot := strings.IndexByte(funcName[lastSlash:], '.') + lastSlash

	return funcName[lastSlash+1 : firstDot]
}
