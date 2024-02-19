package broker

import (
	"fmt"

	"github.com/IBM/sarama"
)

func ConnectConsumer(brokersUrl []string) (sarama.Consumer, error) {

	config := sarama.NewConfig()

	config.Consumer.Return.Errors = true

	config.Consumer.Offsets.AutoCommit.Enable = false

	// NewConsumer creates a new consumer using the given broker addresses and configuration
	conn, err := sarama.NewConsumer(brokersUrl, config)

	if err != nil {
		fmt.Println("Connection failed")
		return nil, err
	}
	return conn, nil
}
