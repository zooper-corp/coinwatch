package blockcypher

type accountResponse struct {
	Address      string `json:"address"`
	FinalBalance int    `json:"final_balance"`
}
