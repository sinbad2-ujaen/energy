package request

type NewTransactionWalletRequest struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	Token float64 `json:"token"`
	Data  string  `json:"data"`
	Type  string  `json:"type"`
}
