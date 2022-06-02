package minaexplorer

import (
	"github.com/zooper-corp/CoinWatch/config"
	"log"
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
	log.Println(r)
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
			Name: "minaexplorer",
		},
		Filters: []config.TokenFilter{
			{
				Symbol:  "mina",
				Address: "B62qq3TQ8AP7MFYPVtMx5tZGF3kWLJukfwG1A1RGvaBW1jfTPTkDBW6",
				Config: config.TokenConfig{
					Symbol:   "mina",
					GeckoId:  "mina-protocol",
					Contract: "mina",
				},
			},
		},
	}
}
