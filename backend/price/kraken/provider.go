package kraken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/scylladb/go-set"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	apiEndpoint = "https://api.kraken.com/0/public/%s?%s"
	apiTicker   = "Ticker"
)

type Provider struct {
	httpClient *http.Client
	builtins   []config.TokenConfig
}

func New(builtins []config.TokenConfig, httpClient *http.Client) Provider {
	return Provider{
		httpClient: httpClient,
		builtins:   builtins,
	}
}

func (p Provider) Name() string {
	return "Kraken"
}

func (p Provider) GetPrices(tokens []string, fiat string) (data.TokenPrices, error) {
	r := make([]data.TokenPrice, 0)
	seen := set.NewStringSet()
	pairs := make([]string, 0)
	for _, token := range tokens {
		if seen.Has(strings.ToUpper(token)) {
			continue
		}
		seen.Add(strings.ToUpper(token))
		pairs = append(pairs, fmt.Sprintf("%s%s", strings.ToUpper(token), strings.ToUpper(fiat)))
	}
	log.Printf("Kraken query prices for: %v", seen)
	pairParam := strings.Join(pairs, ",")
	// Query
	d, err := p.call(apiTicker, fmt.Sprintf("pair=%s", pairParam), nil)
	if err != nil {
		return data.TokenPrices{}, err
	}
	var ticker tickerUnmarshal
	err = json.Unmarshal(d, &ticker)
	if err != nil {
		log.Printf("Unable to unmarshal kraken data: %v\n", err)
		return data.TokenPrices{}, err
	}
	for pair, value := range ticker.Result {
		token := ""
		for _, t := range tokens {
			// BTC special name
			if strings.EqualFold(t, "BTC") && strings.Contains(pair, "XXBT") {
				token = t
				break
			}
			// Others
			if strings.Contains(pair, strings.ToUpper(t)) {
				token = t
				break
			}
		}
		if token == "" {
			continue
		}
		price, err := strconv.ParseFloat(value.B[0], 32)
		if err != nil {
			log.Printf("Unable to decode price from result: %v\n", err)
			return data.TokenPrices{}, err
		}
		log.Printf("Kraken got price for %s => %v", pair, price)
		seen.Remove(strings.ToUpper(token))
		r = append(r, data.TokenPrice{
			Token: token,
			Price: float32(price),
			Fiat:  fiat,
		})
	}
	// Check fiat to fiat
	if seen.Has(strings.ToUpper(fiat)) {
		seen.Remove(strings.ToUpper(fiat))
		r = append(r, data.TokenPrice{
			Token: fiat,
			Price: 1.0,
			Fiat:  fiat,
		})
	}
	// Did we miss anything?
	if seen.Size() > 0 {
		log.Printf("Cannot find price for tokens %v", seen)
		return data.TokenPrices{}, fmt.Errorf("cannot find price for tokens %v", seen)
	}
	// Check if some token was not found
	return data.TokenPrices{Entries: r}, nil
}

func (p Provider) call(method string, params string, data map[string]string) ([]byte, error) {
	uri := fmt.Sprintf(apiEndpoint, method, params)
	if data == nil {
		data = map[string]string{}
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Unable to unrmarshal kraken data: %v\n", err)
		return nil, err
	}
	req, err := http.NewRequest("GET", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Kraken GET query failed: %v\n", req.Response.StatusCode)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	r, err, code, _ := tools.ReadHTTPRequest(req, p.httpClient)
	if err != nil {
		log.Printf("Kraken HTTP request failed: [%d] %v\n", code, err)
		return nil, err
	}
	return r, nil
}
