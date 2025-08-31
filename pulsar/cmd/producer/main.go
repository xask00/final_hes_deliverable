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

	// Create producer for the topic (this will create the topic if it doesn't exist)
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "my-topic",
	})
	if err != nil {
		log.Fatal("Could not instantiate Pulsar producer: ", err)
	}
	defer producer.Close()

	ctx := context.Background()

	// Send messages
	for i := 0; i < 5; i++ {
		message := fmt.Sprintf("Hello Pulsar! Message #%d", i+1)

		_, err := producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: []byte(message),
		})

		if err != nil {
			log.Printf("Failed to send message: %v", err)
		} else {
			fmt.Printf("Sent: %s\n", message)
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("All messages sent successfully!")
}
