package endpoint

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	k "github.com/iaPlotnikovv/myserver/internal/app/init/broker"
	dbase "github.com/iaPlotnikovv/myserver/internal/app/init/database"
)

var msgCount int

var x int64

func Consume() {

	PrintMessage(" Consumer started ")

	topic := "comments"

	worker, err := k.ConnectConsumer([]string{"kafka:9092"})

	if err != nil {
		fmt.Println("NO CONNECTION CONSUMER")
		panic(err)
	}

	// Calling ConsumePartition. It will open one connection per broker
	// and share it for all partitions that live on it.

	consumer, err := worker.ConsumePartition(topic, 0, x)

	defer consumer.Close()

	if err != nil {
		panic(err)
	}

	sigchan := make(chan os.Signal, 1)

	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Get signal for finish
	doneCh := make(chan struct{})

	go func() {

		select {

		case err := <-consumer.Errors():

			fmt.Println(err)

		case msg := <-consumer.Messages():

			msgCount++

			fmt.Printf("Received message Count %d: | Topic(%s) | Message(%s) \n", msgCount, string(msg.Topic), string(msg.Value))

			cmt := string(msg.Value)

			db := dbase.Init()
			// dynamic
			insertDynStmt := `insert into "test"("comment") values($1)`

			_, err_db := db.Exec(insertDynStmt, cmt)

			CheckErr(err_db)

			PrintMessage("Inserting comment into DB")

			fmt.Println("Processed", msgCount, "messages")

			x = consumer.HighWaterMarkOffset()

		case <-sigchan:
			fmt.Println("Interrupt is detected")
			doneCh <- struct{}{}
		}

	}()

	<-doneCh
	if err := worker.Close(); err != nil {
		panic(err)
	}

}
