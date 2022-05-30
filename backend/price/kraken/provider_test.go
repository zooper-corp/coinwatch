package kraken

import (
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"net/http"
	"os"
	"testing"
)

func TestProvider_GetPrices(t *testing.T) {
	provider := New([]config.TokenConfig{{
		Symbol:   "ksm",
		GeckoId:  "kusama",
		Contract: "kusama",
		Decimals: 12,
	}}, http.DefaultClient)
	ps, err := provider.GetPrices([]string{"ksm"}, "usd")
	if err != nil {
		t.Error(err)
	}
	for _, p := range ps.Entries {
		if p.Price == 0.0 {
			t.Errorf("Price is zero for %v", p)
		}
	}
	kp := ps.GetPrice("KSM")
	if kp == 0.0 {
		t.Errorf("Price is zero for GetPrice(KSM)")
	}
	_ = os.Remove(data.GetTestDbPath())
}
