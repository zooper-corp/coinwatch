package display

import (
	"bytes"
	"fmt"
	"github.com/wcharczuk/go-chart"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"math"
	"strings"
	"time"
)

type BmpGraphStyle struct {
	Width  int
	Height int
}

func GetDefaultBmpGraphStyle() BmpGraphStyle {
	return BmpGraphStyle{
		Width:  1024,
		Height: 256,
	}
}

func TotalBmpGraph(c *client.Client, days int, cfg BmpGraphStyle) (*bytes.Buffer, error) {
	bs, err := c.QueryBalance(data.BalanceQueryOptions{Days: days})
	if err != nil {
		return nil, fmt.Errorf("Unable to query balances %v\n", err)
	}
	entries := bs.GetTimeSeries(days, time.Hour*24)
	// Get tokens and sort them by total value
	unsortedTokens := bs.Tokens()
	tokens := make([]string, len(unsortedTokens))
	last := make([]float64, len(unsortedTokens))
	for i, t := range unsortedTokens {
		last[i] = entries[0].FilterToken(t).TotalFiatValue()
	}
	order := tools.SortAndReturnIndex(last)
	for i, idx := range order {
		tokens[i] = unsortedTokens[idx]
	}
	// Create a series for every token
	series := make([]chart.TimeSeries, 0)
	for _, t := range tokens {
		series = append(series, chart.TimeSeries{
			Name:    strings.ToUpper(t),
			Style:   chart.StyleShow(),
			XValues: make([]time.Time, 0),
			YValues: make([]float64, 0),
		})
	}
	// Do axes reverse
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		ts := entry.Entries()[0].Timestamp
		tv := entry.TotalFiatValue()
		// For each token we stack the total value at given time
		for i, t := range tokens {
			series[i].XValues = append(series[i].XValues, ts)
			series[i].YValues = append(series[i].YValues, tv)
			tv -= math.Max(0, entry.FilterToken(t).TotalFiatValue())
		}
	}
	// Draw
	gs := make([]chart.Series, len(series))
	graph := chart.Chart{
		Title: fmt.Sprintf("Total %d days", days),
		TitleStyle: chart.Style{
			Show:     true,
			FontSize: 10,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top: 50,
			},
		},
		Width:  cfg.Width,
		Height: cfg.Height,
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			Style: chart.StyleShow(),
			ValueFormatter: func(v interface{}) string {
				return chart.FloatValueFormatterWithFormat(v, "%.0f")
			},
		},
		Series: gs,
	}
	for i, s := range series {
		s.Style.FillColor = graph.GetColorPalette().GetSeriesColor(i).WithAlpha(50)
		gs[i] = s
	}
	graph.Series = gs
	// Legend
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	// Render
	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	return buffer, err
}
