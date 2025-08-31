package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func main() {
	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		log.Fatal("Could not instantiate Pulsar client: ", err)
	}
	defer client.Close()

	// Create consumer to receive messages
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       "my-topic",
		SubscriptionName:            "my-subscription-" + fmt.Sprintf("%d", time.Now().Unix()),
		SubscriptionInitialPosition: pulsar.SubscriptionPositionEarliest,
	})
	if err != nil {
		log.Fatal("Could not instantiate Pulsar consumer: ", err)
	}
	defer consumer.Close()

	ctx := context.Background()
	fmt.Println("Starting to receive messages... (Press Ctrl+C to stop)")

	// Receive messages in a loop
	for {
		msg, err := consumer.Receive(ctx)
		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			continue
		}

		fmt.Printf("Received: %s\n", string(msg.Payload()))

		// Acknowledge the message
		consumer.Ack(msg)
	}
}
