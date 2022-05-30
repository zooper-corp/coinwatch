package tools

import (
	"strconv"
	"strings"
)

func FmtTokenAmount(value string) float64 {
	if string(value) == "null" || strings.Trim(value, " ") == "" {
		return 0
	}
	if s, err := strconv.ParseFloat(value, 64); err == nil {
		return s
	}
	return 0
}
