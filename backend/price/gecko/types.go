package gecko

type Coin struct {
	Symbol string `db:"symbol"`
	CoinId string `db:"coin_id"`
	Name   string `db:"name"`
}

type CoinList struct {
	Coins []Coin
}
