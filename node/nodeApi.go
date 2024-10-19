package node

import (
	"github.com/gin-gonic/gin"
	"log"
)

func initNodeAPI(port string) {
	log.Println("Starting node api...")
	go func() {

		router := gin.Default()

		router.GET("/node/dag", getDagHandler)

		router.POST("/node/newWallet/:seed", postNewWalletHandler)

		router.GET("/node/selectionTips", getSelectionTipsHandler)

		router.POST("/node/transaction", postNewTransactionHandler)

		router.Run(":" + port)
	}()
}
