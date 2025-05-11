package models

type Account struct {
	MakerCommission  int64           `json:"makerCommission"`
	TakerCommission  int64           `json:"takerCommission"`
	BuyerCommission  int64           `json:"buyerCommission"`
	SellerCommission int64           `json:"sellerCommission"`
	CommissionRates  CommissionRates `json:"commissionRates"`
	CanTrade         bool            `json:"canTrade"`
	CanWithdraw      bool            `json:"canWithdraw"`
	CanDeposit       bool            `json:"canDeposit"`
	UpdateTime       uint64          `json:"updateTime"`
	AccountType      string          `json:"accountType"`
	Balances         []Balance       `json:"balances"`
	Permissions      []string        `json:"permissions"`
	UID              int64           `json:"uid"`
}

type CommissionRates struct {
	Maker  string `json:"maker"`
	Taker  string `json:"taker"`
	Buyer  string `json:"buyer"`
	Seller string `json:"seller"`
}

type Balance struct {
	Asset  string
	Free   string
	Locked string
}
