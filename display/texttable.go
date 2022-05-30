package display

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"sort"
	"strings"
	"time"
)

type TableStyle int8

const (
	Default TableStyle = 0
	Wide               = 1
)

type AsciiTableStyle struct {
	Style   TableStyle
	Borders bool
}

func GetDefaultAsciiTableStyle() AsciiTableStyle {
	return AsciiTableStyle{
		Style:   Default,
		Borders: false,
	}
}

func SummaryAsciiTable(c *client.Client, days int, cfg AsciiTableStyle) (string, error) {
	bs, err := c.QueryBalance(data.BalanceQueryOptions{Days: days})
	if err != nil {
		return "", fmt.Errorf("Unable to query balances %v\n", err)
	}
	// Get entries
	entries := bs.LastSample().GroupBySymbol().Entries()
	ts := time.Now()
	if len(entries) > 0 {
		ts = entries[0].Timestamp
	}
	fmt.Printf("Updated: %v\n", ts.String())
	// Sort by value
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FiatValue > entries[j].FiatValue
	})
	// Create table
	t := table.NewWriter()
	t.SetStyle(getTableStyle(cfg))
	// Header config
	lh := DaysToShortName(days)
	t.AppendHeader(table.Row{"Token", "Address", "Balance", "Price", c.GetFiat(), "1D", lh})
	cc := []table.ColumnConfig{
		{Name: "T"},
		{Name: "Address", Hidden: cfg.Style == Default},
		{Name: "Balance", Hidden: cfg.Style == Default},
		{Name: "Price"},
		{Name: "Value"},
		{Name: "1D"},
		{Name: lh},
	}
	t.SetColumnConfigs(cc)
	// Add rows
	for _, b := range entries {
		t.AppendRow(table.Row{
			// Token
			strings.ToUpper(b.Token),
			// Address
			b.ShortAddr(),
			// Balance
			tools.HumanFloat64(b.Balance),
			// Price
			fmt.Sprintf("%s%s", tools.HumanFloat64(b.PricePerToken()), c.GetFiatSymbol()),
			// Total
			fmt.Sprintf("%d%s", int(b.FiatValue), c.GetFiatSymbol()),
			// Fiat change 1 D
			tools.HumanSignedPercent(bs.FiatValueChange(b.Token, 1)),
			// Fiat change 1 W
			tools.HumanSignedPercent(bs.FiatValueChange(b.Token, days)),
		})
	}
	// Add totals
	t.AppendRow(table.Row{
		// Token
		"Total",
		// Address
		"",
		// Balance
		"",
		// Total
		fmt.Sprintf("%d%s", int(bs.LastSample().TotalFiatValue()), c.GetFiatSymbol()),
		// Fiat change 1 D
		tools.HumanSignedPercent(bs.TotalFiatValueChange(1)),
		// Fiat change 1 W
		tools.HumanSignedPercent(bs.TotalFiatValueChange(days)),
	})
	// Done
	return t.Render(), nil
}

func AllocationAsciiTable(c *client.Client, days int, cfg AsciiTableStyle) (string, error) {
	bs, err := c.QueryBalance(data.BalanceQueryOptions{Days: days})
	if err != nil {
		return "", fmt.Errorf("Unable to query balances %v\n", err)
	}
	// Get entries
	entries := bs.LastSample().GroupBySymbol().Entries()
	ts := time.Now()
	if len(entries) > 0 {
		ts = entries[0].Timestamp
	}
	fmt.Printf("Updated: %v\n", ts.String())
	// Sort by value
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FiatValue > entries[j].FiatValue
	})
	// Create table
	t := table.NewWriter()
	t.SetStyle(getTableStyle(cfg))
	lh := DaysToShortName(days)
	t.AppendHeader(table.Row{"T", "Pct", "Bal", "Price", "1D", lh})
	cc := []table.ColumnConfig{
		{Name: "T"},
		{Name: "Allocation"},
		{Name: "Balance"},
		{Name: "Price"},
		{Name: "1D"},
		{Name: lh},
	}
	t.SetColumnConfigs(cc)
	// Add rows
	for _, b := range entries {
		t.AppendRow(table.Row{
			// Token
			strings.ToUpper(b.Token),
			// Address
			tools.HumanPercent(1.0 / bs.TotalFiatValue() * b.FiatValue),
			// Balance
			tools.HumanFloat64(b.Balance),
			// Price
			fmt.Sprintf("%s%s", tools.HumanFloat64(b.PricePerToken()), c.GetFiatSymbol()),
			// Fiat change 1 D
			tools.HumanSignedPercent(bs.PricePerTokenChange(b.Token, 1)),
			// Fiat change 1 W
			tools.HumanSignedPercent(bs.PricePerTokenChange(b.Token, days)),
		})
	}
	return t.Render(), nil
}

func getTableStyle(cfg AsciiTableStyle) table.Style {
	style := table.StyleLight
	if !cfg.Borders {
		style.Box = table.BoxStyle{
			Left:             "",
			Right:            "",
			PaddingLeft:      "",
			PaddingRight:     " ",
			MiddleSeparator:  "",
			MiddleHorizontal: "",
		}
		style.Options = table.Options{
			DrawBorder:      false,
			SeparateColumns: false,
			SeparateFooter:  false,
			SeparateHeader:  false,
			SeparateRows:    false,
		}
		style.Format = table.FormatOptions{
			Footer: text.FormatTitle,
			Header: text.FormatTitle,
			Row:    text.FormatDefault,
		}
	}
	return style
}
