package tools

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"strconv"
)

func ReverseStringArray(input []string) []string {
	if len(input) == 0 {
		return input
	}
	return append(ReverseStringArray(input[1:]), input[0])
}

func ReverseIntArray(input []int) []int {
	if len(input) == 0 {
		return input
	}
	return append(ReverseIntArray(input[1:]), input[0])
}

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

func NormalizeFloat64Series(series []float64) []float64 {
	r := make([]float64, len(series))
	if len(series) == 0 {
		return r
	}
	max := math.Abs(series[0])
	for _, v := range series {
		max = math.Max(max, math.Abs(v))
	}
	for i, v := range series {
		switch {
		case max > math.Pow10(5):
			rf, err := strconv.ParseFloat(fmt.Sprintf("%.1f", v/1000), 64)
			if err == nil {
				r[i] = rf
			} else {
				r[i] = 0.0
			}
		case max > math.Pow10(2):
			r[i] = float64(int64(v))
		default:
			r[i] = v
		}
	}
	return r
}

func ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}
