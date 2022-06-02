package blockcypher

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
	if r.Balance == 0 && r.Balance < 1 {
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
			Name: "blockcypher",
		},
		Filters: []config.TokenFilter{
			{
				Symbol:  "btc",
				Address: "1DEP8i3QJCsomS4BSMY2RpU1upv62aGvhD",
				Config: config.TokenConfig{
					Symbol:   "btc",
					GeckoId:  "bitcoin",
					Contract: "bitcoin",
				},
			},
		},
	}
}
