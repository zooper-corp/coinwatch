package subscan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	apiEndpoint  = "https://%v.api.subscan.io/api/%v"
	apiTimestamp = "now"
	apiTokens    = "scan/account/tokens"
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
		balance, err := p.GetBalance(f.Config.Contract, f.Address, f.Symbol)
		if err != nil {
			return nil, err
		}
		r = append(r, balance)
	}
	return r, nil
}

func (p Provider) GetBalance(endpoint string, address string, symbol string) (data.TokenBalance, error) {
	r, err := p.call(apiTokens, endpoint, map[string]string{
		"address": address,
	})
	if err != nil {
		return data.TokenBalance{}, err
	}
	var es endpointTokens
	if err := json.Unmarshal(r, &es); err != nil {
		return data.TokenBalance{}, err
	}
	for _, tokenType := range es.Data {
		for _, tb := range tokenType {
			if strings.EqualFold(tb.Symbol, symbol) {
				decimals := tb.Decimals
				balance, _ := tools.ToDecimal(tb.Balance, decimals).Float64()
				locked, _ := tools.ToDecimal(tb.Lock, decimals).Float64()
				log.Printf("Got balance for wallet '%v:%v' => %v/%v", symbol, address, balance, locked)
				return data.TokenBalance{
					Wallet:  p.wallet.Name,
					Symbol:  symbol,
					Address: address,
					Balance: balance,
					Locked:  locked,
				}, nil
			}
		}
	}
	// Empty
	return data.TokenBalance{
		Wallet:  p.wallet.Name,
		Symbol:  symbol,
		Address: address,
		Balance: 0,
		Locked:  0,
	}, nil
}

func (p Provider) Ping(endpoint string) (int, error) {
	r, err := p.call(apiTimestamp, endpoint, nil)
	if err != nil {
		return 0, err
	}
	var et endpointTimestamp
	err = json.Unmarshal(r, &et)
	if err != nil {
		return 0, err
	}
	return et.Data, nil
}

func (p Provider) call(method string, endpoint string, data map[string]string) ([]byte, error) {
	uri := fmt.Sprintf(apiEndpoint, endpoint, method)
	if data == nil {
		data = map[string]string{}
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Unable to unrmarshal subscan data: %v\n", err)
		return nil, err
	}
	for {
		req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Subscan POST query failed: %v\n", req.Response.StatusCode)
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", p.wallet.Provider.Key)
		r, err, code, header := tools.ReadHTTPRequest(req, p.httpClient)
		// Rate limit hit, wait and retry
		if code == 429 {
			retryIn := 2
			// Try to get next retry from headers
			if val, ok := header["Retry-After"]; ok && len(val) > 0 {
				i, err := strconv.Atoi(val[0])
				if err == nil {
					retryIn = i
				}
			}
			// Retry
			log.Printf("Subscan API rate limit exceeded, asked to wait %v seconds\n", retryIn)
			time.Sleep(time.Second * time.Duration(retryIn))
		} else {
			if err != nil {
				log.Printf("Subscan decode response error %v\n", code)
				return nil, err
			}
			return r, err
		}
	}
}
