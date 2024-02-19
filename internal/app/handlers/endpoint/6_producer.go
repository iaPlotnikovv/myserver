package endpoint

import (
	"fmt"

	"github.com/IBM/sarama"
	k "github.com/iaPlotnikovv/myserver/internal/app/init/broker"
)

func PushCommentToQueue(topic string, message []byte) error {

	fmt.Println("producer starts init")

	brokersUrl := []string{"kafka:9092"}

	producer, err := k.ConnectProducer(brokersUrl)

	if err != nil {
		return err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	partition, offset, err := producer.SendMessage(msg)

	if err != nil {
		return err
	}
	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)

	return nil
}
