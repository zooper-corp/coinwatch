package kraken

type tickerUnmarshal struct {
	Error  []string                   `json:"error"`
	Result map[string]resultUnmarshal `json:"result"`
}

type resultUnmarshal struct {
	A []string `json:"a"`
	B []string `json:"b"`
	C []string `json:"c"`
	V []string `json:"v"`
	P []string `json:"p"`
	T []int    `json:"t"`
	L []string `json:"l"`
	H []string `json:"h"`
	O string   `json:"o"`
}
