package pubsub

import (
	"context"
	"encoding/json"
	"energy/domain/entity"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"log"
	"time"
)

var pubsubClient *pubsub.PubSub
var walletCreateTopic *pubsub.Topic
var newTransactionTopic *pubsub.Topic

var hostEntity host.Host

var ctx = context.Background()

var pubsubInput PubsubInputInterface

type PubSubDI struct {
	PubsubOutput PubsubOutputInterface
}

func NewPubSubDI(pubsubOutput PubsubOutputInterface) *PubSubDI {
	return &PubSubDI{
		PubsubOutput: pubsubOutput,
	}
}

func StartPubsub(pubsubInputInterface PubsubInputInterface) {
	log.Println("Initializing pubsub...")

	pubsubInput = pubsubInputInterface

	initInternalPubSub()

	initTopics()
}

func initInternalPubSub() {
	// Init host
	var errHost error
	hostEntity, errHost = libp2p.New()
	if errHost != nil {
		log.Fatal(errHost)
	}
	defer hostEntity.Close()

	log.Printf("Host ID is %s\n", hostEntity.ID())

	// Set up a DHT (Distributed Hash Table) for peer discovery
	kademliaDHT, errDHT := dht.New(ctx, hostEntity)
	if errDHT != nil {
		log.Fatal(errDHT)
	}
	errDHT = kademliaDHT.Bootstrap(ctx)
	if errDHT != nil {
		log.Fatal(errDHT)
	}

	// Init MDNS for peer discovery
	notifee := &CustomNotifee{}
	mdnsService := mdns.NewMdnsService(hostEntity, "Energy", notifee)
	defer mdnsService.Close()

	// Create a Gossipsub instance for pub-sub communication
	var errPubsub error
	pubsubClient, errPubsub = pubsub.NewGossipSub(ctx, hostEntity)
	if errPubsub != nil {
		log.Fatal(errPubsub)
	}

	hostEntity.SetStreamHandler("/energy", func(stream network.Stream) {
		fmt.Fprintln(stream, "Hello, libp2p!")
		stream.Close()
	})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-ticker.C:
				peers := hostEntity.Network().Peers()
				fmt.Println("Connected Peers:")
				for _, p := range peers {
					fmt.Println(p)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func initTopics() {
	log.Println("Initializing pubsub topics...")

	var errTopic error
	walletCreateTopic, errTopic = pubsubClient.Join("wallet-create-topic")
	if errTopic != nil {
		log.Fatal(errTopic)
	}

	subscriber, errSubscriber := walletCreateTopic.Subscribe()
	if errSubscriber != nil {
		panic(errSubscriber)
	}
	subscribe(subscriber, hostEntity.ID())

	newTransactionTopic, errTopic = pubsubClient.Join("new-transaction-topic")
	if errTopic != nil {
		log.Fatal(errTopic)
	}

	subscriber, errSubscriber = newTransactionTopic.Subscribe()
	if errSubscriber != nil {
		panic(errSubscriber)
	}
	subscribe(subscriber, hostEntity.ID())
}

func subscribe(subscriber *pubsub.Subscription, hostID peer.ID) {
	goroutineCtx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()

		for {
			select {
			case <-goroutineCtx.Done():
				log.Println("Goroutine context canceled, stopping subscription.")
				return
			default:
				msg, err := subscriber.Next(goroutineCtx)
				if err != nil {
					panic(err)
				}

				// only consider messages delivered by other peers
				if msg.ReceivedFrom == hostID {
					continue
				}

				var pubsubMessage entity.PubsubMessage

				if err := json.Unmarshal(msg.Data, &pubsubMessage); err != nil {
					log.Println("Error unmarshaling message:", err)
					continue
				}

				switch pubsubMessage.Type {
				case entity.PubSubWalletCreate:
					PubsubInputInterface.WalletCreateMessage(pubsubInput, pubsubMessage)
				case entity.PubSubNewTransaction:
					PubsubInputInterface.NewTransactionMessage(pubsubInput, pubsubMessage)
				default:
					log.Printf("Unknown message type: %s", pubsubMessage.Type)
				}
			}
		}

	}()
}

type PubsubOutputImpl struct {
}

func (e PubsubOutputImpl) BroadcastMessage(otherContext context.Context, message entity.PubsubMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling PubsubMessage:", err)
		return
	}
	switch message.Type {
	case entity.PubSubWalletCreate:
		err = walletCreateTopic.Publish(otherContext, messageBytes)
		if err != nil {
			log.Println("Error publishing message:", err)
		}
	case entity.PubSubNewTransaction:
		err = newTransactionTopic.Publish(otherContext, messageBytes)
		if err != nil {
			log.Println("Error publishing message:", err)
		}
	}
}
