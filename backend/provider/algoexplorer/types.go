package algoexplorer

type accountResponse struct {
	Account struct {
		Address string `json:"address"`
		Amount  int64  `json:"amount"`
	} `json:"account"`
}
