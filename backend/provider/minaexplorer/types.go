package minaexplorer

type accountResponse struct {
	Account struct {
		PublicKey string `json:"publicKey"`
		Balance   struct {
			Total         string      `json:"total"`
			Unknown       string      `json:"unknown"`
			BlockHeight   int         `json:"blockHeight"`
			LockedBalance interface{} `json:"lockedBalance"`
		} `json:"balance"`
	} `json:"account"`
}
