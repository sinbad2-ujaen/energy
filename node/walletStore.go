package node

import (
	"energy/domain/entity"
	"fmt"
	"sync"
)

type WalletStore struct {
	Mu      sync.RWMutex
	Wallets map[string]*entity.Wallet
}

func (ws *WalletStore) SaveWallet(wallet *entity.Wallet) {
	ws.Mu.Lock()
	defer ws.Mu.Unlock()
	ws.Wallets[wallet.Seed] = wallet
}

func (ws *WalletStore) GetWallet(seed string) (*entity.Wallet, bool) {
	ws.Mu.RLock()
	defer ws.Mu.RUnlock()
	wallet, ok := ws.Wallets[seed]
	return wallet, ok
}

func (ws *WalletStore) IncreaseBalance(seed string, token float64) error {
	ws.Mu.Lock()
	defer ws.Mu.Unlock()

	wallet, exists := ws.Wallets[seed]
	if !exists {
		return fmt.Errorf("wallet not found")
	}

	wallet.IncreaseBalance(token)
	ws.Wallets[seed] = wallet

	return nil
}

func (ws *WalletStore) DecreaseBalance(seed string, token float64) error {
	ws.Mu.Lock()
	defer ws.Mu.Unlock()

	wallet, exists := ws.Wallets[seed]
	if !exists {
		return fmt.Errorf("wallet not found")
	}

	err := wallet.DecreaseBalance(token)
	if err != nil {
		return err
	}

	ws.Wallets[seed] = wallet

	return nil
}

func (ws *WalletStore) AddTransactionToHistory(seed string, transaction *entity.Transaction) error {
	ws.Mu.Lock()
	defer ws.Mu.Unlock()

	wallet, exists := ws.Wallets[seed]
	if !exists {
		return fmt.Errorf("wallet not found")
	}

	wallet.AddTransactionToHistory(transaction)
	ws.Wallets[seed] = wallet

	return nil
}
