package entity

type PubsubMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

const (
	PubSubWalletCreate   string = "wallet-create"
	PubSubNewTransaction string = "new-transaction"
)
