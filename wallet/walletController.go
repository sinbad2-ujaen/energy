package wallet

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"energy/domain/entity"
	"energy/domain/entity/request"
	"energy/domain/entity/response"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

var node string

var nodePort string

func StartWallet(walletPort string, _nodePort string) {
	log.Println("Starting wallet api...")

	nodePort = _nodePort

	initWalletAPI(walletPort)
}

func getNodeHandler(c *gin.Context) {
	log.Println("Received get node request")

	node = getRandomNode()

	c.JSON(http.StatusOK, gin.H{
		"node": node,
	})
}

func postNewWalletHandler(c *gin.Context) {

	if node == "" {
		node = getRandomNode()
	}

	seed, err := generateMasterSeed()
	if err != nil {
		log.Println(err.Error(), err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	nodeEndpoint := node + "/node/newWallet/" + seed

	execute, err := http.Post(nodeEndpoint, "application/json", nil)
	if err != nil {
		log.Println("Error making request to node endpoint:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error making request to node endpoint"})
		return
	}
	defer execute.Body.Close()

	if execute.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", execute.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected status code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"seed": seed,
	})
}

func getRandomNode() string {
	return "http://localhost:" + nodePort
}

func generateMasterSeed() (string, error) {
	seedLength := 64

	randomBytes := make([]byte, seedLength)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("error generating random seed: %v", err)
	}

	randomSeed := hex.EncodeToString(randomBytes)

	return randomSeed, nil
}

func doProofOfWork(transactionData string, maxIterations uint64) (uint64, error) {
	var nonce uint64

	// Ensure a non-negative repeat count
	targetPrefixLength := int(math.Max(1, math.Log2(float64(len(transactionData)/10)))) + 1
	targetPrefix := strings.Repeat("0", targetPrefixLength)

	startTime := time.Now()
	for nonce < maxIterations {
		attempt := fmt.Sprintf("%s%d", transactionData, nonce)

		hash := sha256.New()
		hash.Write([]byte(attempt))
		hashBytes := hash.Sum(nil)
		hashString := hex.EncodeToString(hashBytes)

		//log.Printf("Attempt: %s, Hash: %s", attempt, hashString)

		if strings.HasPrefix(hashString, targetPrefix) {
			elapsedTime := time.Since(startTime)
			log.Println("PoW Elapsed time:", elapsedTime)
			return nonce, nil
		}

		nonce++
	}

	return 0, fmt.Errorf("Proof of work failed after %d iterations", maxIterations)
}

func postNewTransactionHandler(c *gin.Context) {
	log.Println("Received new transaction request")
	var transactionRequest request.NewTransactionWalletRequest

	// Parse transaction request
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response body"})
		return
	}

	log.Println("Request Body:", string(body))

	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(&transactionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate transaction request
	if err := validateNewTransactionRequest(&transactionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction request", "details": err.Error()})
		return
	}

	// Get selection tips from node
	if node == "" {
		node = getRandomNode()
	}

	nodeEndpoint := node + "/node/selectionTips"

	selectionTipsResponse, err := http.Get(nodeEndpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform GET selection tips", "detail": err.Error()})
		return
	}
	defer selectionTipsResponse.Body.Close()

	if selectionTipsResponse.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected status code", "detail": err.Error()})
		return
	}

	selectionTipsBody, err := io.ReadAll(selectionTipsResponse.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading selection tips response body", "detail": err.Error()})
		return
	}

	var selectionTipsResponseDto response.SelectionTipsResponse
	err = json.Unmarshal(selectionTipsBody, &selectionTipsResponseDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON response", "detail": err.Error()})
		return
	}

	// Perform PoW if transaction type is standard
	var nonce uint64
	if transactionRequest.Type == entity.TransactionTypeStandard {
		nonce, err = doProofOfWork(transactionRequest.Data, 1000000)
	}

	// Calculate fee if transaction type is fast
	var fee float64 = 0.0
	if transactionRequest.Type == entity.TransactionTypeFast {
		fee = entity.CalculateFee(transactionRequest.Data)
	}

	// Send transaction to node
	newTransaction := entity.NewTransaction(
		transactionRequest.From,
		transactionRequest.To,
		transactionRequest.Token,
		transactionRequest.Data,
		transactionRequest.Type,
		fee,
		nonce)

	newTransactionNodeRequest := request.NewTransactionNodeRequest{Transaction: newTransaction, SelectionTips: selectionTipsResponseDto.SelectionTips}

	nodeEndpoint = node + "/node/transaction"

	newTransactionNodeRequestJSON, err := json.Marshal(newTransactionNodeRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	execute, err := http.Post(nodeEndpoint, "application/json", bytes.NewBuffer(newTransactionNodeRequestJSON))
	if err != nil {
		log.Println("Error making request to node endpoint:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error making request to node endpoint", "detail": err.Error()})
		return
	}
	defer execute.Body.Close()

	if execute.StatusCode != http.StatusOK {
		body, err := io.ReadAll(execute.Body)
		if err != nil {
			log.Println("Error reading response body:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response body"})
			return
		}

		log.Printf("Unexpected status code: %d", execute.StatusCode)
		log.Println("Request Body:", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected status code"})
		return
	}

	// Return transaction to client
	c.JSON(http.StatusOK, gin.H{
		"transaction": newTransaction,
	})

}

func validateNewTransactionRequest(newTransactionRequest *request.NewTransactionWalletRequest) error {

	if !entity.IsValidAddress(newTransactionRequest.From) {
		return errors.New("invalid transaction FROM address")
	}

	if !entity.IsValidAddress(newTransactionRequest.To) {
		return errors.New("invalid transaction TO address")
	}

	if !entity.IsValidTransactionType(newTransactionRequest.Type) {
		return errors.New("invalid transaction type")
	}

	return nil
}
