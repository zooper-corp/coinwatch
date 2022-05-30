package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	apiEndpoint = "https://api.kraken.com/%s"
	apiBalance  = "Balance"
	apiVersion  = "0"
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
	urlPath := fmt.Sprintf("/%s/private/%s", apiVersion, apiBalance)
	d, err := p.call(urlPath, url.Values{})
	if err != nil {
		return nil, err
	}
	var balance balanceUnmarshal
	err = json.Unmarshal(d, &balance)
	if err != nil {
		log.Printf("Unable to unmarshal kraken data: %v\n", err)
		return nil, err
	}
	if len(balance.Error) > 0 {
		log.Printf("Kraken call failed: %v\n", balance.Error)
		return nil, fmt.Errorf("kraken API call failed: %v", balance.Error[0])
	}
	r := make([]data.TokenBalance, 0)
	for token, amount := range balance.Result {
		parts := strings.Split(token, ".")
		// Symbol clean iup
		name := strings.Trim(parts[0], "XZ2")
		if name == "BT" {
			name = "BTC"
		}
		// Get quantity
		qt, err := strconv.ParseFloat(amount, 32)
		if err != nil {
			log.Printf("Unable to unmarshal token quantity: %v => %v %v", token, amount, err)
			return nil, err
		}
		if qt <= 0.0001 {
			continue
		}
		addr := "Funds"
		locked := 0.0
		if len(parts) > 1 {
			switch strings.ToLower(parts[1]) {
			case "s":
				addr = "Staking"
				locked = qt
			case "p":
				addr = "Parachain"
				locked = qt
			default:
				log.Printf("Unknown modifier %v", parts[1])
			}
		}
		log.Printf("Krakeb balance: %v:%v => %v", name, addr, qt)
		r = append(r, data.TokenBalance{
			Wallet:  p.wallet.Name,
			Symbol:  name,
			Address: addr,
			Balance: qt,
			Locked:  locked,
		})
	}
	return r, nil
}

func (p Provider) call(uriPath string, values url.Values) ([]byte, error) {
	uri := fmt.Sprintf(apiEndpoint, uriPath)
	values.Set("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	// Create signature
	secret, _ := base64.StdEncoding.DecodeString(p.wallet.Provider.Secret)
	signature := createSignature(uriPath, values, secret)
	// Create request
	req, err := http.NewRequest("POST", uri, strings.NewReader(values.Encode()))
	if err != nil {
		log.Printf("Kraken POST query failed: %v\n", req.Response.StatusCode)
		return nil, err
	}
	req.Header.Set("API-Key", p.wallet.Provider.Key)
	req.Header.Set("API-Sign", signature)
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Content-Type", " application/x-www-form-urlencoded; charset=utf-8")
	r, err, code, _ := tools.ReadHTTPRequest(req, p.httpClient)
	if err != nil {
		log.Printf("Kraken HTTP request failed: [%d] %v\n", code, err)
		return nil, err
	}
	return r, nil
}

func createSignature(urlPath string, values url.Values, secret []byte) string {
	// See https://www.kraken.com/help/api#general-usage for more information
	shaSum := getSha256([]byte(values.Get("nonce") + values.Encode()))
	macSum := getHMacSha512(append([]byte(urlPath), shaSum...), secret)
	return base64.StdEncoding.EncodeToString(macSum)
}

// getSha256 creates a sha256 hash for given []byte
func getSha256(input []byte) []byte {
	sha := sha256.New()
	sha.Write(input)
	return sha.Sum(nil)
}

// getHMacSha512 creates a hmac hash with sha512
func getHMacSha512(message, secret []byte) []byte {
	mac := hmac.New(sha512.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}
