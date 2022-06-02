package algoexplorer

import (
	"github.com/zooper-corp/CoinWatch/config"
	"net/http"
	"testing"
)

func TestProvider_GetBalance(t *testing.T) {
	p := getProvider()
	b, err := p.GetBalances()
	if err != nil {
		t.Error(err)
	}
	r := b[0]
	if r.Balance == 0 {
		t.Error("Expected >0 got 0")
	}
}

func getProvider() Provider {
	wallet := getWallet()
	return Provider{
		wallet:     &wallet,
		httpClient: http.DefaultClient,
	}
}

func getWallet() config.Wallet {
	return config.Wallet{
		Name: "test",
		Provider: config.ProviderConfig{
			Name: "algoexplorer",
		},
		Filters: []config.TokenFilter{
			{
				Symbol:  "algo",
				Address: "UD33QBPIM4ZO4B2WK5Y5DYT5J5LYY5FA3IF3G4AVYSCWLCSMS5NYDRW6GE",
				Config: config.TokenConfig{
					Symbol:   "algo",
					GeckoId:  "algorand",
					Contract: "algorand",
				},
			},
		},
	}
}
