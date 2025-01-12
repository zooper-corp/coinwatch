package client

import (
	"github.com/scylladb/go-set"
	"github.com/zooper-corp/CoinWatch/backend/price"
	"github.com/zooper-corp/CoinWatch/backend/provider"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"strings"
	"sync"
	"time"
)

type Client struct {
	config config.Config
	db     data.Db
}

func New(configPath string, dbPath string) (Client, error) {
	cfg, err := config.FromFile(configPath)
	if err != nil {
		return Client{}, err
	}
	db, err := data.FromFile(dbPath)
	if err != nil {
		return Client{}, err
	}
	return Client{
		config: cfg,
		db:     db,
	}, err
}

func (c Client) GetFiat() string {
	return c.config.GetFiat()
}

func (c Client) GetFiatSymbol() string {
	return c.config.GetFiatSymbol()
}

// GetLastBalanceUpdate return timestamp of last balance update
func (c Client) GetLastBalanceUpdate() time.Time {
	last := c.GetLastBalance()
	if len(last.Entries()) > 0 {
		return last.Entries()[0].Timestamp
	}
	return time.UnixMilli(0)
}

// GetLastBalance return last balance series
func (c Client) GetLastBalance() data.Balances {
	b, err := c.db.GetBalances(data.BalanceQueryOptions{Days: 7})
	if err != nil {
		return data.Balances{}
	}
	return b.LastSample()
}

// QueryBalance will fetch data from the DB
func (c Client) QueryBalance(options data.BalanceQueryOptions) (data.Balances, error) {
	return c.db.GetBalances(options)
}

// Get balance within range
func (c Client) GetBalancesFromDate(from time.Time) (data.Balances, error) {
	return c.db.GetBalancesFromDate(from)
}

// UpdateBalance will update the balance for each wallet if rs exceeds updateTtlSeconds
func (c Client) UpdateBalance(updateTtlSeconds int64) error {
	start := time.Now()
	// We are not updating below 5 seconds
	if updateTtlSeconds < 5 {
		updateTtlSeconds = 5
	}
	// Check wallets
	wallets := c.config.GetWallets()
	if len(wallets) == 0 {
		log.Fatalf("No wallet configured")
	}
	// Get current balances, we must update daily so just get last day results
	balances, err := c.db.GetBalances(data.BalanceQueryOptions{Days: 1})
	if err != nil {
		log.Fatalf("Unable to get balances from DB: %v", err)
	}
	// Check if update is required
	updateRequired := false
	for _, wallet := range wallets {
		wb := balances.LastSample().FilterByWallet(wallet.Name).Entries()
		switch {
		case len(wb) == 0:
			log.Printf("Update required, wallet '%v' never updated", wallet.Name)
			updateRequired = true
			break
		case len(wallet.Filters) > 0 && len(wb) != len(wallet.Filters):
			log.Printf("Update required, wallet '%v' has different filters", wallet.Name)
			updateRequired = true
			break
		case start.UnixMilli()-wb[0].Timestamp.UnixMilli() > (updateTtlSeconds * 1000):
			log.Printf("Update required, wallet '%v' update expired", wallet.Name)
			updateRequired = true
			break
		default:
			log.Printf("Skipping wallet '%v', last update less than %d secs ago", wallet.Name, updateTtlSeconds)
		}
	}
	// Update balances if needed
	ch := make(chan tools.Result[[]data.TokenBalance])
	var wg sync.WaitGroup
	if updateRequired {
		for _, wallet := range wallets {
			wg.Add(1)
			w := wallet
			go func() {
				defer wg.Done()
				ch <- tools.ResultFrom(c.updateWallet(&w))
			}()
		}
	}
	// Collect
	go func() {
		wg.Wait()
		close(ch)
		log.Printf("Updated balances in %.2fsecs\n", float64(time.Now().UnixMilli()-start.UnixMilli())/1000.0)
	}()
	updatedBalances := make([]data.TokenBalance, 0)
	tokens := set.NewStringSet()
	for r := range ch {
		if r.IsErr() {
			log.Printf("One provider failed, aborting update")
			return r.Err
		}
		for _, b := range r.Value {
			if strings.EqualFold(c.GetFiat(), b.Symbol) {
				continue
			}
			if !tokens.Has(strings.ToLower(b.Symbol)) {
				tokens.Add(strings.ToLower(b.Symbol))
			}
			updatedBalances = append(updatedBalances, b)
		}
	}
	// No balance
	if len(updatedBalances) == 0 {
		log.Printf("No balances updated")
		return nil
	}
	// Update prices
	log.Println("Updating prices")
	priceProvider := price.New(c.config.GetTokenConfigs(), c.db, c.config.GetHttpClient())
	prices, err := priceProvider.GetPrices(tokens.List(), c.config.GetFiat())
	if err != nil {
		return err
	}
	// Update DB, our TS is our ID
	ts := start.Truncate(time.Second)
	for _, b := range updatedBalances {
		p := 1.0
		if !strings.EqualFold(b.Symbol, c.GetFiat()) {
			p = prices.GetPrice(b.Symbol)
		}
		value := b.Balance * p
		if float32(value) > c.config.GetFiatMin() {
			entry := data.Balance{
				Timestamp:     ts,
				Wallet:        b.Wallet,
				Token:         b.Symbol,
				Address:       b.Address,
				Balance:       b.Balance,
				BalanceLocked: b.Locked,
				FiatValue:     value,
			}
			err := c.db.InsertBalance(entry)
			if err != nil {
				return err
			}
		}
	}
	// Done
	return nil
}

func (c *Client) updateWallet(wallet *config.Wallet) ([]data.TokenBalance, error) {
	bp, err := provider.New(wallet, c.config.GetHttpClient())
	if err != nil {
		log.Printf("Cannot get balance provider for wallet %v\n", wallet.Name)
		return nil, err
	}
	// Update data, get balance and current fiat value
	log.Printf("Updating wallet %v from %s\n", wallet.Name, wallet.Provider.Name)
	tb, err := bp.GetBalances()
	if err != nil {
		log.Printf("Error wallet %v from %s: %s\n", wallet.Name, wallet.Provider.Name, err.Error())
	}
	return tb, err
}
