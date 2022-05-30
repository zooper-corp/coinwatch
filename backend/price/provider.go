package price

import (
	"fmt"
	"github.com/scylladb/go-set"
	"github.com/zooper-corp/CoinWatch/backend/price/gecko"
	"github.com/zooper-corp/CoinWatch/backend/price/kraken"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"log"
	"net/http"
	"strings"
)

type Provider interface {
	GetPrices(tokens []string, fiat string) (data.TokenPrices, error)
	Name() string
}

type MultiSourceProvider struct {
	providers []Provider
}

func (p MultiSourceProvider) Name() string {
	return "MultiSource"
}

func (p MultiSourceProvider) GetPrices(tokens []string, fiat string) (data.TokenPrices, error) {
	missing := set.NewStringSet()
	for _, t := range tokens {
		missing.Add(strings.ToUpper(t))
	}
	result := data.TokenPrices{}
	for _, provider := range p.providers {
		if missing.Size() == 0 {
			break
		}
		tp, err := provider.GetPrices(missing.List(), fiat)
		if err != nil {
			log.Printf("Unable to check prices from %v: %v\n", provider.Name(), err)
		} else {
			for _, rtp := range tp.Entries {
				missing.Remove(strings.ToUpper(rtp.Token))
				result.Entries = append(result.Entries, rtp)
			}
		}
	}
	if missing.Size() > 0 {
		return result, fmt.Errorf("unable to load prices for %v", missing.List())
	}
	return result, nil
}

func New(builtins []config.TokenConfig, db data.Db, httpClient *http.Client) Provider {
	cg := gecko.New(builtins, db)
	k := kraken.New(builtins, httpClient)
	return MultiSourceProvider{[]Provider{cg, k}}
}
