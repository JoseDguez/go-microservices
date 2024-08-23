package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/JoseDguez/go-microservices/internal/ledger"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	dbDriver = "mysql"
	dbName   = "ledger"
	topic    = "ledger"
	groupID  = "ledger-consumer-group"
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

var db *sql.DB

func main() {
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	var err error

	dbUser := os.Getenv("MYSQL_USERNAME")
	dbPassword := os.Getenv("MYSQL_PASSWORD")

	// Connect to the database
	dsn := fmt.Sprintf("%s:%s@tcp(mysql-ledger:3306)/%s", dbUser, dbPassword, dbName)

	db, err = sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = db.Close(); err != nil {
			log.Printf("Failed to close database connection: %s", err)
		}
	}()

	// Ping the database to ensure the connection is active
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

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
	var ledgerMsg LedgerMsg
	err := json.Unmarshal(msg.Value, &ledgerMsg)
	if err != nil {
		log.Println("Failed to unmarshal message in handleMessage:", err)
		return
	}

	err = ledger.Insert(db, ledgerMsg.OrderID, ledgerMsg.UserID, ledgerMsg.Amount, ledgerMsg.Operation, ledgerMsg.Date)
	if err != nil {
		log.Println("Failed to insert data in handleMessage:", err)
		return
	}
}
