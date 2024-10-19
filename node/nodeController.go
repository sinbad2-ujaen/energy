package node

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"energy/domain/entity"
	"energy/domain/entity/request"
	"energy/pubsub"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"strings"
)

var walletStore = NewWalletStore()

var pendingTransactions = NewPendingTransactions()

var pubsubOutput pubsub.PubsubOutputInterface

type Node struct {
	PubsubInput pubsub.PubsubInputInterface
}

func NewNodeDI(pubsubInput pubsub.PubsubInputInterface) *Node {
	return &Node{
		PubsubInput: pubsubInput,
	}
}

func StartNode(pubsubOutputImpl pubsub.PubsubOutputImpl, port string) {
	log.Println("Initializing node...")

	pubsubOutput = pubsubOutputImpl

	initDAG()
	addPendingTransactionsToDAG()
	initNodeAPI(port)
}

func NewWalletStore() *WalletStore {
	walletStore := &WalletStore{
		Wallets: make(map[string]*entity.Wallet),
	}

	// Origin wallet
	originWallet := entity.NewWallet("9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04")
	originWallet.IncreaseBalance(1000000000)
	walletStore.SaveWallet(originWallet)

	return walletStore
}

func NewPendingTransactions() *entity.PendingTransactions {
	pendingTransactions := &entity.PendingTransactions{
		Transactions: make(map[string]entity.Transaction),
	}

	return pendingTransactions
}

func addPendingTransactionsToDAG() {
	// Fake method to add pending transactions to the DAG
	pendingTransactions.InitPendingTransactions()

	for _, transaction := range pendingTransactions.Transactions {
		dag.addTransaction(transaction)
	}
}

func sendWalletCreatePubSubMessage(ctx context.Context, a_wallet *entity.Wallet) {
	walletJson, err := json.Marshal(a_wallet)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	pubsubOutput.BroadcastMessage(ctx, entity.PubsubMessage{
		Type: entity.PubSubWalletCreate,
		Data: string(walletJson),
	})
}

func getDagHandler(c *gin.Context) {
	c.JSON(http.StatusOK, dag)
}

func postNewWalletHandler(c *gin.Context) {

	walletSeed := c.Param("seed")

	if walletSeed == "" {
		log.Println("Invalid wallet seed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet seed"})
		return
	}

	newWallet := entity.NewWallet(walletSeed)
	walletStore.SaveWallet(newWallet)

	sendWalletCreatePubSubMessage(c, newWallet)

	c.JSON(http.StatusOK, gin.H{
		"seed": walletSeed,
	})
}

func getSelectionTipsHandler(c *gin.Context) {
	transactions := pendingTransactions.GetOldestTransactions()

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}

func postNewTransactionHandler(c *gin.Context) {
	var newTransactionRequest request.NewTransactionNodeRequest

	if err := c.ShouldBindJSON(&newTransactionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateTransaction(&newTransactionRequest.Transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateSelectionTips(&newTransactionRequest.SelectionTips); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newTransactionRequest.Transaction.UpdateTimestampAdded()

	pendingTransactions.AddTransaction(newTransactionRequest.Transaction)

	dag.addTransaction(newTransactionRequest.Transaction)

	newTransactionNodeRequest := request.NewTransactionNodeRequest{Transaction: newTransactionRequest.Transaction, SelectionTips: newTransactionRequest.SelectionTips}

	sendNewTransactionPubSubMessage(c, &newTransactionNodeRequest)

	c.JSON(http.StatusOK, gin.H{
		"transaction": newTransactionRequest.Transaction,
	})
}

func sendNewTransactionPubSubMessage(ctx context.Context, newTransactionNodeRequest *request.NewTransactionNodeRequest) {
	newTransactionPubsubJson, err := json.Marshal(newTransactionNodeRequest)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	pubsubOutput.BroadcastMessage(ctx, entity.PubsubMessage{
		Type: entity.PubSubNewTransaction,
		Data: string(newTransactionPubsubJson),
	})
}

func validateTransaction(newTransaction *entity.Transaction) error {
	if !entity.IsValidID(newTransaction.ID) {
		return errors.New("invalid transaction id")
	}

	if !entity.IsValidAddress(newTransaction.From) {
		return errors.New("invalid transaction FROM address")
	}

	if !entity.IsValidAddress(newTransaction.To) {
		return errors.New("invalid transaction TO address")
	}

	if !entity.IsValidTransactionType(newTransaction.Type) {
		return errors.New("invalid transaction type")
	}

	if newTransaction.Type == entity.TransactionTypeFast {
		if !entity.IsValidFee(newTransaction.Fee, newTransaction.Data) {
			return errors.New("invalid transaction fee")
		}
		fromWallet, exists := walletStore.GetWallet(newTransaction.From)
		if !exists {
			return errors.New("not enough tokens to perform transaction")
		}

		if newTransaction.Fee+newTransaction.Token > fromWallet.Balance {
			return errors.New("not enough tokens to perform transaction")
		}
	}

	if newTransaction.Type == entity.TransactionTypeStandard {
		if !verifyProofOfWork(newTransaction.Data, newTransaction.Nonce) {
			return errors.New("invalid PoW")
		}

		if newTransaction.Token > 0 {
			fromWallet, exists := walletStore.GetWallet(newTransaction.From)
			if !exists {
				return errors.New("not enough tokens to perform transaction")
			}

			if newTransaction.Token > fromWallet.Balance {
				return errors.New("not enough tokens to perform transaction")
			}
		}

	}
	return nil
}

func validateSelectionTips(selectionTips *[]entity.Transaction) error {
	for _, selectionTip := range *selectionTips {

		if err := validateTransaction(&selectionTip); err != nil {
			return errors.New("invalid selection tip. error: " + err.Error())
		}

		if err := dag.ConfirmTransaction(selectionTip.ID); err != nil {
			return errors.New("error confirming transaction")
		}

		tokenInterchangeErr := performTokenInterchange(selectionTip.From, selectionTip.To, selectionTip.Token, selectionTip.Fee)
		if tokenInterchangeErr != nil {
			return tokenInterchangeErr
		}

		pendingTransactions.DeleteTransaction(selectionTip.ID)

		//Add transaction to wallet history
		selectionTipWallet, exists := walletStore.GetWallet(selectionTip.From)
		if !exists {
			return errors.New("from wallet not found")
		}
		selectionTipWallet.AddTransactionToHistory(&selectionTip)
	}

	return nil
}

func verifyProofOfWork(transactionData string, nonce uint64) bool {
	targetPrefixLength := int(math.Max(1, math.Log2(float64(len(transactionData)/10)))) + 1
	targetPrefix := strings.Repeat("0", targetPrefixLength)

	// Combine transaction data with nonce
	attempt := fmt.Sprintf("%s%d", transactionData, nonce)

	// Calculate hash
	hash := sha256.New()
	hash.Write([]byte(attempt))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	// Check if the hash meets the PoW criteria
	return hashString[:len(targetPrefix)] == targetPrefix
}

type PubsubInputImpl struct {
}

func (e PubsubInputImpl) WalletCreateMessage(message entity.PubsubMessage) {
	fmt.Printf("Node received message: %s\n", message.Data)

	var newWallet *entity.Wallet
	err := json.Unmarshal([]byte(message.Data), newWallet)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	_, exists := walletStore.GetWallet(newWallet.Seed)
	if !exists {
		walletStore.SaveWallet(newWallet)
	}
}

func (e PubsubInputImpl) NewTransactionMessage(message entity.PubsubMessage) {
	fmt.Printf("Node received message: %s\n", message.Data)

	var newTransactionNodeRequest request.NewTransactionNodeRequest
	err := json.Unmarshal([]byte(message.Data), &newTransactionNodeRequest)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	if err := validateTransaction(&newTransactionNodeRequest.Transaction); err != nil {
		fmt.Println("Error validating transaction:", err)
		return
	}

	if err := validateSelectionTips(&newTransactionNodeRequest.SelectionTips); err != nil {
		fmt.Println("Error validating selection tips:", err)
		return
	}

	pendingTransactions.AddTransaction(newTransactionNodeRequest.Transaction)

	dag.addTransaction(newTransactionNodeRequest.Transaction)

	fmt.Printf("Node added transaction: %s\n", newTransactionNodeRequest.Transaction.ID)
}

func performTokenInterchange(from string, to string, token float64, fee float64) error {
	fromWallet, exists := walletStore.GetWallet(from)
	if !exists {
		return errors.New("from wallet not found performing token interchange")
	}

	toWallet, exists := walletStore.GetWallet(to)
	if !exists {
		toWallet = entity.NewWallet(to)
		walletStore.SaveWallet(toWallet)
	}

	errDecreasingBalance := walletStore.DecreaseBalance(fromWallet.Seed, token+fee)
	if errDecreasingBalance != nil {
		return errDecreasingBalance
	}
	errIncreasingBalance := walletStore.IncreaseBalance(toWallet.Seed, token)
	if errIncreasingBalance != nil {
		// If increasing balance fails, restore the balance of the sender
		restatingBalanceErr := walletStore.IncreaseBalance(fromWallet.Seed, token+fee)
		if restatingBalanceErr != nil {
			return restatingBalanceErr
		}
		return errIncreasingBalance
	}

	return nil
}
