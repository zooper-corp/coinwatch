package data

import (
	"fmt"
	"github.com/scylladb/go-set"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"strings"
	"time"
)

type Balance struct {
	Timestamp     time.Time `db:"ts"`
	Wallet        string    `db:"wallet"`
	Token         string    `db:"token"`
	Address       string    `db:"address"`
	Balance       float64   `db:"balance"`
	BalanceLocked float64   `db:"balance_locked"`
	FiatValue     float64   `db:"fiat_value"`
}

type TokenBalance struct {
	Wallet  string
	Symbol  string
	Address string
	Balance float64
	Locked  float64
}

type TokenPrice struct {
	Token string
	Price float32
	Fiat  string
}

type TokenPrices struct {
	Entries []TokenPrice
}

type Balances struct {
	entries []Balance
}

// ShortAddr returns the address
func (b Balance) ShortAddr() string {
	l := len(b.Address)
	if l < 12 {
		return b.Address
	}
	return fmt.Sprintf("%s...%s", b.Address[0:4], b.Address[l-5:l-1])
}

// Id is a unique key for this balance
func (b Balance) Id() string {
	return fmt.Sprintf("%v/%v/%v", b.Wallet, b.Token, b.Address)
}

// PricePerToken returns price per token in fiat value
func (b Balance) PricePerToken() float64 {
	return b.FiatValue / b.Balance
}

// Add creates a new balance with a sum of the two
func (b Balance) Add(x Balance) Balance {
	return Balance{
		Timestamp:     b.Timestamp,
		Wallet:        "Grouped",
		Token:         b.Token,
		Address:       "Grouped",
		Balance:       b.Balance + x.Balance,
		BalanceLocked: b.BalanceLocked + x.BalanceLocked,
		FiatValue:     b.FiatValue + x.FiatValue,
	}
}

// GetTimeSeries will return a set of balances over a given amount of intervals
func (b Balances) GetTimeSeries(amount int, interval time.Duration) []Balances {
	r := make([]Balances, 0)
	duration := time.Duration(0)
	for i := 1; i < amount; i++ {
		r = append(r, b.ClosestSample(duration))
		duration = duration + interval
	}
	return r
}

// TotalFiatValueChange will return total fiat value change in pct (-1.0 to 1.0) between now and x days ago
func (b Balances) TotalFiatValueChange(days int) float64 {
	startTotal := b.LastSample().TotalFiatValue()
	endTotal := b.ClosestSample(time.Hour * time.Duration(24*days)).TotalFiatValue()
	if startTotal == endTotal {
		return 0
	}
	return (1.0 / endTotal * startTotal) - 1
}

// FiatValueChange will return fiat value change in pct (-1.0 to 1.0) between now and x days ago for token
func (b Balances) FiatValueChange(token string, days int) float64 {
	tuple := b.ClosestSampleTuple(time.Duration(0), time.Hour*time.Duration(24*days)).
		GroupBySymbol().
		FilterToken(token).
		entries
	if len(tuple) < 2 {
		return 0
	}
	startValue := tuple[0].FiatValue
	endValue := tuple[len(tuple)-1].FiatValue
	return (1.0 / endValue * startValue) - 1
}

// PricePerTokenChange will return token price value change in pct (-1.0 to 1.0) between now and x days ago for token
func (b Balances) PricePerTokenChange(token string, days int) float64 {
	tuple := b.ClosestSampleTuple(time.Duration(0), time.Hour*time.Duration(24*days)).
		GroupBySymbol().
		FilterToken(token).
		entries
	if len(tuple) < 2 {
		return 0
	}
	startPrice := tuple[0].PricePerToken()
	endPrice := tuple[len(tuple)-1].PricePerToken()
	return (1.0 / endPrice * startPrice) - 1
}

// BalanceChange will return change for balance with given id in pct (-1.0 to 1.0) between now and x days ago
func (b Balances) BalanceChange(id string, days int) float64 {
	tuple := b.ClosestSampleTuple(time.Duration(0), time.Hour*time.Duration(24*days)).
		GroupBySymbol().
		FilterId(id).
		entries
	if len(tuple) < 2 {
		return 0
	}
	startBalance := tuple[0].Balance
	endBalance := tuple[len(tuple)-1].Balance
	return (1.0 / endBalance * startBalance) - 1
}

// FilterId will return entries filtered by id
func (b Balances) FilterId(id string) Balances {
	if len(b.entries) == 0 {
		return Balances{}
	}
	r := make([]Balance, 0)
	for _, bs := range b.entries {
		if bs.Id() == id {
			r = append(r, bs)
		}
	}
	return Balances{entries: r}
}

// FilterToken will return entries filtered by token
func (b Balances) FilterToken(token string) Balances {
	if len(b.entries) == 0 {
		return Balances{}
	}
	r := make([]Balance, 0)
	for _, bs := range b.entries {
		if strings.EqualFold(bs.Token, token) {
			r = append(r, bs)
		}
	}
	return Balances{entries: r}
}

// Tokens will return all tokens in this series
func (b Balances) Tokens() []string {
	if len(b.entries) == 0 {
		return nil
	}
	r := set.NewStringSet()
	for _, bs := range b.entries {
		if !r.Has(strings.ToUpper(bs.Token)) {
			r.Add(strings.ToUpper(bs.Token))
		}
	}
	return r.List()
}

// Wallets will return all tokens in this series
func (b Balances) Wallets() []string {
	if len(b.entries) == 0 {
		return nil
	}
	r := set.NewStringSet()
	for _, bs := range b.entries {
		if !r.Has(bs.Wallet) {
			r.Add(bs.Wallet)
		}
	}
	return r.List()
}

// ClosestSampleTuple return 2 samples between start and end
func (b Balances) ClosestSampleTuple(start time.Duration, end time.Duration) Balances {
	r := make([]Balance, 0)
	if startSeries := b.ClosestSample(start).entries; len(startSeries) > 0 {
		for _, bs := range startSeries {
			r = append(r, bs)
		}
	}
	if endSeries := b.ClosestSample(end).entries; len(endSeries) > 0 {
		for _, bs := range endSeries {
			r = append(r, bs)
		}
	}
	return Balances{entries: r}
}

// LastSample will return the last sample
func (b Balances) LastSample() Balances {
	return b.ClosestSample(time.Duration(0))
}

// ClosestSample will find the sample closes to duration from now
func (b Balances) ClosestSample(duration time.Duration) Balances {
	if len(b.entries) == 0 {
		return Balances{}
	}
	targetTs := time.Now().Add(-duration)
	bestDelta := tools.AbsDuration(targetTs.Sub(b.entries[0].Timestamp))
	bestTs := b.entries[0].Timestamp
	// Find closest TS first
	for _, be := range b.entries {
		beDelta := tools.AbsDuration(targetTs.Sub(be.Timestamp))
		// We are closer, update bestDelta
		if beDelta < bestDelta {
			bestDelta = beDelta
			bestTs = be.Timestamp
		}
	}
	// Append series
	r := make([]Balance, 0)
	for _, be := range b.entries {
		if be.Timestamp.UnixMilli() == bestTs.UnixMilli() {
			r = append(r, be)
		}
	}
	// Done
	return Balances{entries: r}
}

// GroupBySymbol will group values by symbol
func (b Balances) GroupBySymbol() Balances {
	if len(b.entries) == 0 {
		return Balances{}
	}
	r := make([]Balance, 0)
	ts := time.UnixMilli(0)
	br := make(map[string]Balance, 0)
	for _, bs := range b.entries {
		// TS Changed
		if bs.Timestamp.UnixMilli() != ts.UnixMilli() {
			if len(br) > 0 {
				for _, v := range br {
					r = append(r, v)
				}
			}
			br = make(map[string]Balance, 0)
			ts = bs.Timestamp
		}
		// Map
		token := strings.ToUpper(bs.Token)
		v, ok := br[token]
		if ok {
			br[token] = v.Add(bs)
		} else {
			bCopy := bs
			br[token] = bCopy
		}
	}
	// Add last
	for _, v := range br {
		r = append(r, v)
	}
	// Done
	return Balances{entries: r}
}

func (b Balances) FilterByWallet(name string) Balances {
	r := make([]Balance, 0)
	for _, balance := range b.entries {
		if strings.EqualFold(balance.Wallet, name) {
			r = append(r, balance)
		}
	}
	return Balances{entries: r}
}

func (b Balances) Entries() []Balance {
	a := b.entries
	return a
}

func (b Balances) TotalFiatValue() float64 {
	total := 0.0
	seen := set.NewStringSet()
	for _, be := range b.entries {
		key := be.Id()
		if !seen.Has(key) {
			seen.Add(key)
			total += be.FiatValue
		}
	}
	return total
}

// GetPrice returns price for a given token or 0 if not found
func (tp *TokenPrices) GetPrice(token string) float64 {
	for _, p := range tp.Entries {
		if strings.EqualFold(p.Token, token) {
			return float64(p.Price)
		}
	}
	log.Printf("Price not found %s", token)
	return 0
}
