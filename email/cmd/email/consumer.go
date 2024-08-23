package main

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/JoseDguez/go-microservices/internal/email"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	topic   = "email"
	groupID = "email-consumer-group"
)

type ConsumerGroupHandler struct{}

func (ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Partition %d - Received message: %s\n", message.Partition, string(message.Value))
		handleMessage(message)
		session.MarkMessage(message, "")
	}
	return nil
}

func main() {
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumerGroup, err := sarama.NewConsumerGroup([]string{"my-cluster-kafka-bootstrap:9092"}, groupID, saramaConfig)
	if err != nil {
		log.Fatalf("Failed to create sarama consumer group: %v", err)
	}
	defer func(consumerGroup sarama.ConsumerGroup) {
		err := consumerGroup.Close()
		if err != nil {
			log.Println("Failed to close sarama consumer group:", err)
		}
	}(consumerGroup)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, ConsumerGroupHandler{}); err != nil {
				log.Fatalf("Error from consumer: %v", err)
			}
			// Check if context was cancelled, signaling that we should stop consuming
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Handle termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Received shutdown signal")
	cancel()
	wg.Wait()
}

func handleMessage(msg *sarama.ConsumerMessage) {
	var emailMsg EmailMsg
	err := json.Unmarshal(msg.Value, &emailMsg)
	if err != nil {
		log.Println("Failed to unmarshal message in handleMessage:", err)
		return
	}

	err = email.Send(emailMsg.UserID, emailMsg.OrderId)
	if err != nil {
		log.Println("Failed to send email in handleMessage:", err)
		return
	}
}
