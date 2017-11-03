package utils

import (
	"strconv"
	"strings"
)

func MustAtoi(s string) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		panic(err)
	}
	return i
}

func MustAtof64(s string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		panic(err)
	}
	return f
}
