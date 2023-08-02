package config

import "time"

type configUnmarshal struct {
	Globals globals       `yaml:"globals"`
	Wallets []wallet      `yaml:"wallets"`
	Tokens  []TokenConfig `yaml:"tokens"`
}

type TelegramBotConfig struct {
	ApiToken       string
	ChatId         int64
	UpdateInterval time.Duration
}

type Config struct {
	globals globals
	wallets []wallet
	tokens  []TokenConfig
}

type globals struct {
	Fiat       string  `yaml:"fiat"`
	FiatSymbol string  `yaml:"fiat_symbol"`
	FiatMin    float32 `yaml:"fiat_min"`
}

type wallet struct {
	Name     string         `yaml:"name"`
	Provider ProviderConfig `yaml:"provider"`
	Tokens   []string       `yaml:"tokens"`
}

type Wallet struct {
	Name     string
	Provider ProviderConfig
	Filters  []TokenFilter
}

type TokenFilter struct {
	Symbol  string
	Address string
	Config  TokenConfig
}

type ProviderConfig struct {
	Name   string   `yaml:"name"`
	Key    string   `yaml:"key"`
	Secret string   `yaml:"secret"`
	Ignore []string `yaml:"ignore"`
}

type TokenConfig struct {
	Symbol   string `yaml:"symbol"`
	GeckoId  string `yaml:"geckoid"`
	Contract string `yaml:"contract"`
}
