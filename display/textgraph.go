package display

import (
	"fmt"
	"github.com/guptarohit/asciigraph"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"time"
)

type AsciiGraphStyle struct {
	Width  int
	Height int
}

func GetDefaultAsciiGraphStyle() AsciiGraphStyle {
	return AsciiGraphStyle{
		Width:  80,
		Height: 10,
	}
}

func TotalAsciiGraph(c *client.Client, days int, cfg AsciiGraphStyle) (string, error) {
	bs, err := c.QueryBalance(data.BalanceQueryOptions{Days: days})
	if err != nil {
		return "", fmt.Errorf("Unable to query balances %v\n", err)
	}
	series := bs.GetTimeSeries(cfg.Width, time.Hour*time.Duration(days)/time.Duration(cfg.Width))
	d := make([]float64, 0)
	for _, entry := range series {
		d = append(d, entry.TotalFiatValue())
	}
	d = tools.NormalizeFloat64Series(d)
	tools.ReverseFloat64Array(d)
	return asciigraph.Plot(
		d,
		asciigraph.Height(cfg.Height),
		asciigraph.Width(cfg.Width),
	), nil
}
