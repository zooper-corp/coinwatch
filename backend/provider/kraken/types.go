package kraken

type balanceUnmarshal struct {
	Error  []string          `json:"error"`
	Result map[string]string `json:"result"`
}
