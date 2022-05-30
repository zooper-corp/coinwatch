package subscan

type endpointTimestamp struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	GeneratedAt int    `json:"generated_at"`
	Data        int    `json:"data"`
}

type endpointSearch struct {
	Code        int `json:"code"`
	GeneratedAt int `json:"generated_at"`
	Data        struct {
		Account struct {
			Balance     string `json:"balance"`
			BalanceLock string `json:"balance_lock"`
		} `json:"account"`
	} `json:"data"`
}
