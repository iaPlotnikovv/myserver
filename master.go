package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
)

const (
	host     = "postgres"
	port     = 5432
	user     = "postgres"
	password = "test"
	dbname   = "mydb"
)

// инициализируем соединение с БД

func Init() *sql.DB {

	var err error

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		fmt.Printf("Ошибка, %s", err)
	}

	err = db.Ping()

	if err != nil {
		fmt.Printf("Ошибка ping, %s", err)
	}
	return db
}

// ошибки:

func checkErr(err error) {
	if err != nil {
		fmt.Printf("Ошибка, %s", err)
		panic(err)
	}
}

//----------------------------------

func main() {

	//create mux
	mux := http.NewServeMux()

	mux.HandleFunc("/", empty)

	pHandler := Ilia{}
	mux.Handle("/plotnikov", pHandler)

	mux.HandleFunc("/plotnikov/db", GetInfo)

	mux.HandleFunc("/plotnikov/db_post", PostInfo)

	//server

	s := &http.Server{
		Addr:    ":1311",
		Handler: mux,
	}

	s.ListenAndServe()

}

//-----------------------------------------------------------

func empty(res http.ResponseWriter, req *http.Request) {

	data := []byte("Welcome! try /plotnikov!")
	res.WriteHeader(200)
	res.Write(data)
}

type Ilia struct{}

func (p Ilia) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	data := []byte("HELLO WORLD! I'm Ilia!")
	res.WriteHeader(200)
	res.Write(data)
}

//curl -v -X GET http://localhost:1311/plotnikov

func printMessage(message string) {
	fmt.Println("")
	fmt.Println(message)
	fmt.Println("")
}

// -------------------------------------------
type info_js struct {
	ID      int    `json:"id"`
	Comment string `json:"comment"`
}

type JsonResponse struct {
	Type string    `json:"type"`
	Data []info_js `json:"data"`
}

// Fetch db
func GetInfo(res http.ResponseWriter, req *http.Request) {

	db := Init()

	printMessage("Getting info...")

	// Get all  from  table
	rows, err := db.Query("SELECT * FROM test")

	checkErr(err)

	// var response []JsonResponse
	var info []info_js

	for rows.Next() {
		snb := info_js{}
		err := rows.Scan(&snb.ID, &snb.Comment)
		if err != nil {
			fmt.Println(err)
			http.Error(res, http.StatusText(500), 500)
			return
		}
		info = append(info, snb)
	}

	if err = rows.Err(); err != nil {
		http.Error(res, http.StatusText(500), 500)
		return
	}
	var response = JsonResponse{Type: "success", Data: info}

	json.NewEncoder(res).Encode(response)

	// loop and display the result in the browser
	fmt.Fprintf(res, "\nId | comment")
	fmt.Fprintf(res, "\n------------\n")

	for _, snb := range info {
		fmt.Fprintf(res, "%d  |  %s\n\n", snb.ID, snb.Comment)
	}
}

// POST!
func PostInfo(w http.ResponseWriter, r *http.Request) {

	comment := r.FormValue("comment")
	mesg, _ := json.Marshal(comment)

	var response = JsonResponse{}

	if comment == "" {
		response = JsonResponse{Type: "error"}
	} else {
		PushCommentToQueue("comments", mesg)
		go consume()
	}

	json.NewEncoder(w).Encode(response)
}

// .................Producer........................
//
// Инициализируем соединение с кафкой продюсером
func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	fmt.Println("producer trying to connect")
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	//NewSyncProducer creates a new SyncProducer using the given broker addresses and configuration.
	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		printMessage("Connection failed")
		return nil, err
	}
	return conn, nil
}

// Отправляем сообщение в топик кафки
func PushCommentToQueue(topic string, message []byte) error {
	fmt.Println("producer starts init")
	brokersUrl := []string{"kafka:9092"}
	producer, err := ConnectProducer(brokersUrl)
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

//...........................................

// ..............Consumer..................
// Коннектимся к кафке консъюмером
func connectConsumer(brokersUrl []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	// NewConsumer creates a new consumer using the given broker addresses and configuration
	conn, err := sarama.NewConsumer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func consume() {

	printMessage("Consumer started ")

	topic := "comments"
	worker, err := connectConsumer([]string{"kafka:9092"})
	if err != nil {
		panic(err)
	}

	// Calling ConsumePartition. It will open one connection per broker
	// and share it for all partitions that live on it.

	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	sigchan := make(chan os.Signal, 1)

	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Count how many message processed
	msgCount := 0

	// Get signal for finish
	doneCh := make(chan struct{})

	go func() {
		for {
			select {
			case err := <-consumer.Errors():

				fmt.Println(err)

			case msg := <-consumer.Messages():

				msgCount++

				fmt.Printf("Received message Count %d: | Topic(%s) | Message(%s) \n", msgCount, string(msg.Topic), string(msg.Value))

				cmt := string(msg.Value)

				db := Init()
				// dynamic
				insertDynStmt := `insert into "test"("comment") values($1)`

				_, err_db := db.Exec(insertDynStmt, cmt)

				checkErr(err_db)

				printMessage("Inserting comment into DB")

				consumer.Pause()

			case <-sigchan:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")

	if err := worker.Close(); err != nil {
		panic(err)
	}

}

//...........................................
