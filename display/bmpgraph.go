package display

import (
	"bytes"
	"fmt"
	"github.com/wcharczuk/go-chart"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/data"
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
	series := bs.GetTimeSeries(days, time.Hour*24)
	// Create axes
	xAxis := make([]time.Time, 0)
	yAxis := make([]float64, 0)
	for i := len(series) - 1; i >= 0; i-- {
		entry := series[i]
		ts := entry.Entries()[0].Timestamp
		// Create bar
		xAxis = append(xAxis, ts)
		yAxis = append(yAxis, entry.TotalFiatValue())
	}
	// Draw
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
			Style:          chart.StyleShow(),
			ValueFormatter: chart.FloatValueFormatter,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "Totals",
				Style:   chart.StyleShow(),
				XValues: xAxis,
				YValues: yAxis,
			},
		},
	}
	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	return buffer, err
}
