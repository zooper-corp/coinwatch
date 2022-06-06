package algoexplorer

import (
	"encoding/json"
	"fmt"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"net/http"
)

const (
	apiEndpoint = "https://algoindexer.algoexplorerapi.io/v2/%v"
	apiAccount  = "accounts/"
)

type Provider struct {
	wallet     *config.Wallet
	httpClient *http.Client
}

func New(wallet *config.Wallet, httpClient *http.Client) (Provider, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return Provider{
		wallet:     wallet,
		httpClient: httpClient,
	}, nil
}

func (p Provider) GetBalances() ([]data.TokenBalance, error) {
	r := make([]data.TokenBalance, 0)
	for _, f := range p.wallet.Filters {
		balance, err := p.GetBalance(f.Address, f.Symbol)
		if err != nil {
			return nil, err
		}
		r = append(r, balance)
	}
	return r, nil
}

func (p Provider) GetBalance(address string, symbol string) (data.TokenBalance, error) {
	r, err := p.call(apiAccount + address)
	print(string(r))
	if err != nil {
		return data.TokenBalance{}, err
	}
	var account accountResponse
	err = json.Unmarshal(r, &account)
	if err != nil {
		return data.TokenBalance{}, err
	}
	return data.TokenBalance{
		Wallet:  p.wallet.Name,
		Symbol:  symbol,
		Address: account.Account.Address,
		Balance: float64(account.Account.Amount) / 1000000.0,
		Locked:  0,
	}, nil
}

func (p Provider) call(uriPath string) ([]byte, error) {
	uri := fmt.Sprintf(apiEndpoint, uriPath)
	// Create request
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Printf("Algo Explorer create query failed: %v\n", req.Response.StatusCode)
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	)
	req.Header.Set("Content-Type", "application/json")
	r, err, code, _ := tools.ReadHTTPRequest(req, p.httpClient)
	if err != nil {
		log.Printf("Algo Explorer HTTP request failed: [%d] %v\n", code, err)
		return nil, err
	}
	return r, nil
}
