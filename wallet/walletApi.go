package wallet

import (
	"github.com/gin-gonic/gin"
	"log"
)

func initWalletAPI(port string) {
	log.Println("Starting wallet api...")
	go func() {

		router := gin.Default()

		router.GET("/wallet/getNode", getNodeHandler)

		router.POST("/wallet", postNewWalletHandler)

		router.POST("/wallet/transaction", postNewTransactionHandler)

		router.Run(":" + port)
	}()
}
