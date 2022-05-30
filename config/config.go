package config

import (
	"embed"
	"fmt"
	"github.com/zooper-corp/CoinWatch/tools"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strings"
)

//go:embed tokens/*.yml
var sourcesConfig embed.FS

func FromFile(path string) (Config, error) {
	data, err := ioutil.ReadFile(tools.ExpandPath(path))
	if err != nil {
		return Config{}, err
	}
	return FromData(data)
}

func FromData(data []byte) (Config, error) {
	var config configUnmarshal
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}
	// Add builtin sources
	builtins, err := sourcesConfig.ReadDir("tokens")
	if err != nil {
		return Config{}, err
	}
	for _, bi := range builtins {
		data, err := sourcesConfig.ReadFile(fmt.Sprintf("tokens/%v", bi.Name()))
		if err != nil {
			return Config{}, err
		}
		var tokens []TokenConfig
		if err := yaml.Unmarshal(data, &tokens); err != nil {
			return Config{}, err
		}
		for _, t := range tokens {
			config.Tokens = append(config.Tokens, t)
		}
	}
	// Done
	return Config{
		globals: config.Globals,
		wallets: config.Wallets,
		tokens:  config.Tokens,
	}, nil
}

func (c *Config) GetTokenConfigs() []TokenConfig {
	return c.tokens
}

func (c *Config) GetHttpClient() *http.Client {
	return http.DefaultClient
}

func (c *Config) GetFiat() string {
	return c.globals.Fiat
}

func (c *Config) GetFiatMin() float32 {
	return c.globals.FiatMin
}

func (c *Config) GetFiatSymbol() string {
	return c.globals.FiatSymbol
}

func (c *Config) GetWallets() []Wallet {
	r := make([]Wallet, 0)
	for _, w := range c.wallets {
		filters := make([]TokenFilter, 0)
		for _, t := range w.Tokens {
			ts := strings.Split(strings.Trim(t, " "), ":")
			filter := TokenFilter{
				Symbol: strings.ToLower(ts[0]),
			}
			// Address is optional
			if len(ts) > 1 {
				filter.Address = ts[1]
			}
			for _, tc := range c.tokens {
				if strings.EqualFold(tc.Symbol, filter.Symbol) {
					filter.Config = tc
					break
				}
			}
			filters = append(filters, filter)
		}
		r = append(r, Wallet{
			Name:     w.Name,
			Provider: w.Provider,
			Filters:  filters,
		})
	}
	return r
}
