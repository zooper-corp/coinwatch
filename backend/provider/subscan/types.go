package subscan

type endpointTimestamp struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	GeneratedAt int    `json:"generated_at"`
	Data        int    `json:"data"`
}

type endpointTokenData struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	Balance  string `json:"balance"`
	Lock     string `json:"lock"`
}

type endpointTokens struct {
	Code        int                            `json:"code"`
	Message     string                         `json:"message"`
	GeneratedAt int                            `json:"generated_at"`
	Data        map[string][]endpointTokenData `json:"data"`
}
