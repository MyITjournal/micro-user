package consumer

import (
	"context"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/uloamaka/notification-service/email_worker/processor"
)

type KafkaEmailConsumer struct {
	processor *processor.EmailProcessor
}

func NewKafkaEmailConsumer(p *processor.EmailProcessor) *KafkaEmailConsumer {
	return &KafkaEmailConsumer{processor: p}
}

func (c *KafkaEmailConsumer) Start(topic string) error {

	kafkaServer := os.Getenv("KAFKA_BROKER")
	if kafkaServer == "" {
		kafkaServer = "kafka:29092"
	}
	log.Printf("🔹 Kafka broker being used: %s\n", kafkaServer)

	kafkaGroup := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroup == "" {
		kafkaGroup = "email-worker"
	}

	conf := &kafka.ConfigMap{
		"bootstrap.servers":  kafkaServer,
		"group.id":           kafkaGroup,
		"auto.offset.reset":  "earliest",
	}

	consumer, err := kafka.NewConsumer(conf)
	if err != nil {
		return err
	}
	defer consumer.Close()

	topicName := topic
	if topicName == "" {
		topicName = os.Getenv("KAFKA_TOPIC")
		if topicName == "" {
			topicName = "email_service"
		}
	}

	if err := consumer.Subscribe(topicName, nil); err != nil {
		return err
	}

	log.Printf("📩 Kafka consumer connected to %s, subscribed to topic: %s\n", kafkaServer, topicName)

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("⚠️ Error reading message: %v\n", err)
			continue
		}

		go c.processor.Process(context.Background(), msg.Value)
	}
}
