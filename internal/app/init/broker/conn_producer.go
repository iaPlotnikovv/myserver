package broker

import (
	"fmt"

	"github.com/IBM/sarama"
)

func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {

	fmt.Println("producer trying to connect")

	config := sarama.NewConfig()

	config.Producer.Return.Successes = true

	config.Producer.RequiredAcks = sarama.WaitForAll

	config.Producer.Retry.Max = 5
	//NewSyncProducer creates a new SyncProducer using the given broker addresses and configuration.

	conn, err := sarama.NewSyncProducer(brokersUrl, config)

	if err != nil {
		fmt.Println("Connection failed")
		return nil, err
	}
	return conn, nil
}
