package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

const KafkaTopic = "email.jobs"

// ✅ The EXACT struct your email worker consumes
type EmailJob struct {
	UserID     string `json:"user_id"`
	To         string `json:"to"`
	Subject    string `json:"subject"`
	HtmlBody   string `json:"html"`
	TextBody   string `json:"text"`
	TemplateID string `json:"template_id"`
	RequestID  string `json:"request_id"`
}

func main() {
	// Create Kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		log.Fatal("Failed to create producer:", err)
	}
	defer p.Close()

	job := EmailJob{
		UserID:     uuid.New().String(),
		To:         "fivivi2625@canvect.com",
		Subject:    "Kafka Email Test",
		HtmlBody:   "<h1>Hello from Kafka!</h1><p>Your email worker works ✅</p>",
		TextBody:   "Hello from Kafka! Your email worker works.",
		TemplateID: "test_template",
		RequestID:  uuid.New().String(),
	}

	value, err := json.Marshal(job)
	if err != nil {
		log.Fatal("Failed to marshal JSON:", err)
	}

	// Send message
	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &[]string{KafkaTopic}[0], Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)

	if err != nil {
		log.Fatal("Failed to produce:", err)
	}

	p.Flush(5000)

	fmt.Println("✅ Email job produced successfully!")
}
