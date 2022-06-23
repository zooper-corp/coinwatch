package subscan

import (
	"github.com/zooper-corp/CoinWatch/config"
	"net/http"
	"testing"
)

func TestProvider_Ping(t *testing.T) {
	p := getProvider(getPolkadotWallet())
	r, err := p.Ping("polkadot")
	if err != nil {
		t.Error(err)
	}
	if r == 0 {
		t.Error("Expected >0 got 0")
	}
}

func TestProvider_GetBalance(t *testing.T) {
	p := getProvider(getPolkadotWallet())
	b, err := p.GetBalances()
	if err != nil {
		t.Error(err)
	}
	r := b[0]
	if r.Balance == 0 {
		t.Error("Expected >0 got 0")
	}
}

func TestProvider_Erc20_GetBalance(t *testing.T) {
	p := getProvider(getErc20MoonBeamWallet())
	b, err := p.GetBalances()
	if err != nil {
		t.Error(err)
	}
	r := b[0]
	if r.Balance == 0 {
		t.Error("Expected >0 got 0")
	}
}

func getProvider(wallet config.Wallet) Provider {
	return Provider{
		wallet:     &wallet,
		httpClient: http.DefaultClient,
	}
}

func getPolkadotWallet() config.Wallet {
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
				},
			},
		},
	}
}

func getErc20MoonBeamWallet() config.Wallet {
	return config.Wallet{
		Name: "test",
		Provider: config.ProviderConfig{
			Name: "subscan",
		},
		Filters: []config.TokenFilter{
			{
				Symbol:  "well",
				Address: "0x519ee031E182D3E941549E7909C9319cFf4be69a",
				Config: config.TokenConfig{
					Symbol:   "well",
					GeckoId:  "moonwell",
					Contract: "moonbeam",
				},
			},
		},
	}
}
