package provider

import (
	"github.com/zooper-corp/CoinWatch/backend/provider/algoexplorer"
	"github.com/zooper-corp/CoinWatch/backend/provider/blockcypher"
	"github.com/zooper-corp/CoinWatch/backend/provider/kraken"
	"github.com/zooper-corp/CoinWatch/backend/provider/minaexplorer"
	"github.com/zooper-corp/CoinWatch/backend/provider/subscan"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"log"
	"net/http"
	"strings"
)

type Provider interface {
	GetBalances() ([]data.TokenBalance, error)
}

func New(wallet *config.Wallet, httpClient *http.Client) (Provider, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	switch strings.ToLower(wallet.Provider.Name) {
	case "subscan":
		return subscan.New(wallet, httpClient)
	case "algoexplorer":
		return algoexplorer.New(wallet, httpClient)
	case "blockcypher":
		return blockcypher.New(wallet, httpClient)
	case "minaexplorer":
		return minaexplorer.New(wallet, httpClient)
	case "kraken":
		return kraken.New(wallet, httpClient)
	default:
		log.Fatalf("Invalid balance provider %v\n", wallet.Provider)
	}
	return nil, nil
}
