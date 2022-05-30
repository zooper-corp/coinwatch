package tools

import (
	"fmt"
	"math"
	"strconv"
)

func ReverseFloat64Array(a []float64) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func HumanSignedPercent(x float64) string {
	v := x * 100
	sign := '+'
	if v < 0 {
		sign = '-'
	}
	return fmt.Sprintf("%c%s", sign, HumanPercent(x))
}

func HumanPercent(x float64) string {
	v := x * 100
	a := math.Abs(v)
	switch {
	case a == math.Round(x):
		return fmt.Sprintf("%.0f%%", a)
	case a > 10:
		return fmt.Sprintf("%.0f%%", a)
	case a > 1:
		return fmt.Sprintf("%.1f%%", a)
	case a < 0.1:
		return "0%"
	default:
		return fmt.Sprintf("%.1f%%", a)
	}
}

func HumanFloat64(v float64) string {
	switch {
	case v == math.Round(v):
		return fmt.Sprintf("%.0f", v)
	case v > math.Pow10(6):
		return fmt.Sprintf("%.1fM", v/math.Pow10(6))
	case v > math.Pow10(3):
		return fmt.Sprintf("%.1fK", v/math.Pow10(3))
	case v > math.Pow10(2):
		return fmt.Sprintf("%.0f", v)
	case v > math.Pow10(1):
		return fmt.Sprintf("%.1f", v)
	case v > math.Pow10(1):
		return fmt.Sprintf("%.2f", v)
	case v < 1:
		return fmt.Sprintf("%.3f", v)
	default:
		return fmt.Sprintf("%.2f", v)
	}
}

func NormalizeFloat64(v float64) float64 {
	switch {
	case v > math.Pow10(5):
		r, err := strconv.ParseFloat(fmt.Sprintf("%.1f", v/1000), 64)
		if err == nil {
			return r
		} else {
			return 0.0
		}
	case v > math.Pow10(2):
		return float64(int64(v))
	default:
		return v
	}
}
