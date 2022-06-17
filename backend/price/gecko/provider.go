package gecko

import (
	gecko "github.com/superoo7/go-gecko/v3"
	"github.com/superoo7/go-gecko/v3/types"
	"github.com/upper/db/v4"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"log"
	"strings"
)

type Provider struct {
	client   *gecko.Client
	builtins []config.TokenConfig
	db       data.Db
}

func New(builtins []config.TokenConfig, db data.Db) Provider {
	client := gecko.NewClient(nil)
	return Provider{
		client:   client,
		builtins: builtins,
		db:       db,
	}
}

func (cg Provider) Name() string {
	return "CoinGecko"
}

func (cg Provider) GetPrices(tokens []string, fiat string) (data.TokenPrices, error) {
	vc := []string{fiat}
	coins, err := cg.getCoinList(tokens)
	if err != nil {
		return data.TokenPrices{}, err
	}
	sp, err := cg.client.SimplePrice(coins.GetTokens(), vc)
	if err != nil {
		return data.TokenPrices{}, err
	}
	log.Printf("CoinGecko prices for tokens %v -> r: %v", coins.GetTokens(), sp)
	prices := make([]data.TokenPrice, 0)
	for _, coin := range coins.Coins {
		price := (*sp)[coin.CoinId][strings.ToLower(fiat)]
		if price != 0 {
			prices = append(prices, data.TokenPrice{
				Token: coin.Symbol,
				Price: price,
				Fiat:  strings.ToLower(fiat),
			})
		}
	}
	return data.TokenPrices{Entries: prices}, nil
}

func (cl CoinList) GetTokens() []string {
	var r = make([]string, 0)
	for _, c := range cl.Coins {
		r = append(r, c.CoinId)
	}
	return r
}

func (cg Provider) getCoinList(tokens []string) (CoinList, error) {
	result := make([]Coin, 0)
	// Check builtins
	for _, tg := range cg.builtins {
		// No gecko id provided
		if strings.Trim(tg.GeckoId, " ") == "" {
			continue
		}
		// Check if requested
		for i, t := range tokens {
			if strings.EqualFold(t, tg.Symbol) {
				result = append(result, Coin{
					CoinId: tg.GeckoId,
					Name:   tg.GeckoId,
					Symbol: tg.Symbol,
				})
				// Delete without preserving order
				tokens[i] = tokens[len(tokens)-1]
				tokens = tokens[:len(tokens)-1]
				break
			}
		}
	}
	// No token left
	if len(tokens) == 0 {
		return CoinList{Coins: result}, nil
	}
	// Check DB now
	sess, err := cg.db.GetSession()
	if err != nil {
		return CoinList{}, err
	}
	defer func(sess db.Session) {
		_ = sess.Close()
	}(sess)
	collection := sess.Collection("gecko_coins")
	// Update collection if needed
	exists, _ := collection.Exists()
	if !exists {
		log.Printf("Create gecko coin map cache table")
		_, err = sess.SQL().Exec(`
        CREATE TABLE gecko_coins (
            symbol TEXT,
			name TEXT,
			coin_id TEXT
        )`)
		if err != nil {
			log.Fatalf("Unable to create coins gecko cache table")
			return CoinList{}, err
		}
	}
	// Look for coins
	var coinList types.CoinList
	for _, token := range tokens {
		r := collection.Find("symbol", strings.ToLower(token))
		exists, err = r.Exists()
		if err != nil || !exists {
			if coinList == nil {
				log.Printf("Fetching symbols: %v\n", tokens)
				list, err := cg.client.CoinsList()
				if err != nil {
					log.Fatalf("Unable to load data from coin gecko")
					return CoinList{}, err
				}
				coinList = *list
			}
			found := false
			for _, c := range coinList {
				if strings.EqualFold(c.Symbol, token) {
					found = true
					coin := Coin{
						CoinId: c.ID,
						Name:   c.Name,
						Symbol: c.Symbol,
					}
					_, err := collection.Insert(coin)
					if err != nil {
						log.Fatalf("Unable to insert coin %v", c)
						return CoinList{}, err
					}
					result = append(result, coin)
					break
				}
			}
			if !found {
				log.Fatalf("Unable to find token %v on coin gecko", token)
				return CoinList{}, err
			}
		}
		// Coin is present in DB
		var coin Coin
		err = r.One(&coin)
		if err != nil {
			return CoinList{}, err
		}
		result = append(result, coin)
	}
	// All good
	log.Printf("CoinGecko symbols: %v", result)
	return CoinList{Coins: result}, nil
}
