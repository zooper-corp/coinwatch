package subscan

import (
	"github.com/zooper-corp/CoinWatch/config"
	"net/http"
	"testing"
)

func TestProvider_Ping(t *testing.T) {
	p := getProvider()
	r, err := p.Ping("polkadot")
	if err != nil {
		t.Error(err)
	}
	if r == 0 {
		t.Error("Expected >0 got 0")
	}
}

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
			Name: "subscan",
		},
		Filters: []config.TokenFilter{
			{
				Symbol:  "dot",
				Address: "1vTfju3zruADh7sbBznxWCpircNp9ErzJaPQZKyrUknApRu",
				Config: config.TokenConfig{
					Symbol:   "dot",
					GeckoId:  "polkadot",
					Contract: "polkadot",
					Decimals: 12,
				},
			},
		},
	}
}
