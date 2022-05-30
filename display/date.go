package display

import "fmt"

func DaysToShortName(days int) string {
	r := fmt.Sprintf("%vD", days)
	switch days {
	case 7:
		r = "1W"
	case 31:
		r = "1M"
	case 365:
		r = "1Y"
	}
	return r
}
