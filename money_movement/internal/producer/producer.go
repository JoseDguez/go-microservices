package producer

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
	"sync"
	"time"
)

const (
	emailTopic  = "email"
	ledgerTopic = "ledger"
)

type EmailMsg struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

type LedgerMsg struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	Operation string `json:"operation"`
	Date      string `json:"date"`
}

func SendCaptureMessage(pid string, userID string, amount int64) {
	producer, err := sarama.NewSyncProducer([]string{"my-cluster-kafka-bootstrap:9092"}, sarama.NewConfig())
	if err != nil {
		log.Println("Failed to start Sarama producer in SendCaptureMessage:", err)
		return
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Println("Failed to close Sarama producer in SendCaptureMessage:", err)
		}
	}()

	emailMsg := EmailMsg{
		OrderID: pid,
		UserID:  userID,
	}

	ledgerMsg := LedgerMsg{
		OrderID:   pid,
		UserID:    userID,
		Amount:    amount,
		Operation: "DEBIT",
		Date:      time.Now().Format("2006-01-02"),
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go sendMsg(producer, emailMsg, emailTopic, &wg)
	go sendMsg(producer, ledgerMsg, ledgerTopic, &wg)
	wg.Wait()
}

func sendMsg[T EmailMsg | LedgerMsg](producer sarama.SyncProducer, msg T, topic string, wg *sync.WaitGroup) {
	stringMsg, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal message in sendMsg:", err)
		return
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(stringMsg),
	}

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		log.Println("Failed to send message in sendMsg:", err)
		return
	}

	log.Printf("Message sent to partition %d at offset %d\n", partition, offset)

	wg.Done()
}
