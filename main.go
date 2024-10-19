package main

import (
	"context"
	"energy/node"
	"energy/pubsub"
	"energy/wallet"
	"fmt"
	"log"
	"os"
	"os/signal"
)

var ctx, cancel = context.WithCancel(context.Background())

func main() {
	defer cancel()

	var nodePort string
	var walletPort string
	if len(os.Args) > 2 {
		nodePort = os.Args[1]
		walletPort = os.Args[2]
	} else {
		log.Printf("No ports specified. Using default ports %s", "8080 / 8090")
		nodePort = "8080"
		walletPort = "8090"
	}

	log.Println("Initializing Energy...")

	pubsubInputImpl := node.PubsubInputImpl{}

	pubsubOutputImpl := pubsub.PubsubOutputImpl{}

	// Node DI & Start

	node.NewNodeDI(pubsubInputImpl)

	node.StartNode(pubsubOutputImpl, nodePort)

	// Pubsub DI & Start

	pubsub.NewPubSubDI(pubsubOutputImpl)

	pubsub.StartPubsub(pubsubInputImpl)

	// Wallet Start

	wallet.StartWallet(walletPort, nodePort)

	handleExit()
}

func handleExit() {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	select {
	case <-interruptChan:
		fmt.Println("Received interrupt signal. Shutting down...")
	case <-ctx.Done():
		fmt.Println("Shutting down...")
	}
}
