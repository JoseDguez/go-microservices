package email

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/JoseDguez/go-microservices/internal/email"
	"log"
	"sync"
)

const topic = "email"

var wg sync.WaitGroup

type EmailMsg struct {
	OrderId string `json:"order_id"`
	UserID  string `json:"user_id"`
}

func main() {
	done := make(chan struct{})

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, sarama.NewConfig())
	if err != nil {
		log.Fatalf("Failed to start Sarama consumer: %v", err)
	}

	defer func() {
		close(done)
		if err := consumer.Close(); err != nil {
			log.Println("Failed to close Sarama consumer:", err)
		}
	}()

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
	}

	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("Failed to start consumer for partition %d: %v", partition, err)
		}

		defer func() {
			if err := partitionConsumer.Close(); err != nil {
				log.Println("Failed to close Sarama consumer:", err)
			}
		}()

		wg.Add(1)
		go awaitMessages(partitionConsumer, partition, done)
	}

	wg.Wait()
}

func awaitMessages(pc sarama.PartitionConsumer, partition int32, done chan struct{}) {
	defer wg.Done()

	for {
		select {
		case msg := <-pc.Messages():
			log.Printf("Partition %d - Received message: %s\n", partition, string(msg.Value))
			handleMessage(msg)
		case <-done:
			log.Printf("Received done signal. Exiting partition %d\n", partition)
			return
		}
	}
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
