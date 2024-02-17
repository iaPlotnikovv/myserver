package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
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

// инициализируем соединение с Redis

func InitRed() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "my-password", // password set
		DB:       0,             // use default DB
	})

	return client
}

// ----------------------------------
type IliaDB struct{}
type IliaPOST struct{}

func main() {

	//create mux
	mux := http.NewServeMux()

	mux.HandleFunc("/", empty)

	pHandler := IliaDB{}
	POSTHandler := IliaPOST{}

	mux.Handle("/plotnikov", pHandler)

	mux.HandleFunc("/plotnikov/db", FetchInfo)

	mux.Handle("/plotnikov/db_post", POSTHandler)

	mux.HandleFunc("/plotnikov/db_post/", PostInfo)

	//server

	s := &http.Server{
		Addr:    ":1311",
		Handler: mux,
	}

	s.ListenAndServe()

}

//-----------------------------------------------------------

func empty(res http.ResponseWriter, req *http.Request) {

	data := []byte("HELLO WORLD! I'm Ilia!\n Welcome! try /plotnikov!")
	res.WriteHeader(200)
	res.Write(data)
}

func (p IliaDB) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	http.ServeFile(res, req, "db.html")
}
func (p IliaPOST) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	http.ServeFile(res, req, "post.html")

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

// POST!
func PostInfo(w http.ResponseWriter, r *http.Request) {

	//http.ServeFile(w, r, "post.html")

	comment := r.FormValue("comment")
	mesg := []byte(comment)

	var response = JsonResponse{}

	if comment != "" {

		PushCommentToQueue("comments", mesg)
		//time.Sleep(1 * time.Second)
		consume()

	} else {
		response = JsonResponse{Type: "error"}
	}

	json.NewEncoder(w).Encode(response)
}

// display db

func FetchInfo(w http.ResponseWriter, r *http.Request) {

	var rows *sql.Rows
	var err error

	zapis := r.FormValue("comment")

	//zapros := fmt.Sprintf("SELECT id, comment FROM test WHERE comment LIKE '%s'", zapis)

	//dataInRedis, err := redClient.Get(ctx, zapis).Result()

	if zapis != "" {

		response, err := cacheme(zapis)

		if err != nil {

			fmt.Fprintf(w, err.Error()+"\r\n")

		} else {

			fmt.Fprintf(w, "Search result for %s:\n\n", zapis)

			enc := json.NewEncoder(w)

			enc.SetIndent("", "  ")

			if err := enc.Encode(response); err != nil {
				fmt.Println(err.Error())
			}

		}

		printMessage("Getting info...")

	} else {

		db := Init()

		rows, err = db.Query("SELECT * FROM test")

		checkErr(err)

		printMessage("This is DataBase...")

		fmt.Fprintf(w, "DATABASE:\n")

		var info []info_js

		for rows.Next() {

			snb := info_js{}

			err := rows.Scan(&snb.ID, &snb.Comment)

			if err != nil {
				fmt.Println(err)
				http.Error(w, http.StatusText(500), 500)
				return
			}

			info = append(info, snb)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		var response = JsonResponse{Type: "success", Data: info}

		json.NewEncoder(w).Encode(response)

		// loop and display the result in the browser

		fmt.Fprintf(w, "\nId | comment")
		fmt.Fprintf(w, "\n------------\n")

		for _, snb := range info {
			fmt.Fprintf(w, "%d  |  %s\n\n", snb.ID, snb.Comment)
		}
	}

}

func cacheme(zapis string) (*JsonResponse, error) {

	ctx := context.Background()

	client := InitRed()

	cachedComments, err := client.Get(ctx, zapis).Bytes()

	response := JsonResponse{}

	if err != nil {

		dbComments, err := FetchFromDB(zapis)

		if err != nil {
			return nil, err
		}

		cachedComments, err = json.Marshal(dbComments)

		if err != nil {
			return nil, err
		}

		err = client.Set(ctx, zapis, cachedComments, 2*time.Minute).Err()

		response = JsonResponse{Type: "PostgreSQL", Data: dbComments}

		printMessage("FROM PostgreSQL...")

		return &response, err
	}

	comments := []info_js{}

	err = json.Unmarshal(cachedComments, &comments)

	if err != nil {
		return nil, err
	}

	response = JsonResponse{Type: "Redis Cache", Data: comments}

	printMessage("FROM Redis Cache...")

	return &response, nil
}

func FetchFromDB(zapis string) ([]info_js, error) {

	db := Init()

	queryString := fmt.Sprintf("SELECT id, comment FROM test WHERE comment LIKE '%s'", zapis)

	rows, err := db.Query(queryString)

	if err != nil {
		return nil, err
	}

	var info []info_js

	for rows.Next() {

		p := info_js{}
		err = rows.Scan(&p.ID, &p.Comment)

		info = append(info, p)

		if err != nil {
			return nil, err
		}

	}

	return info, nil
}

//..................K A F K A.....................

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
	config.Consumer.Offsets.AutoCommit.Enable = false
	//config.Consumer.Offsets.Initial = sarama.OffsetOldest
	//config.Consumer.Offsets.Retry.Max = 5
	//config.Consumer.Interceptors =

	// NewConsumer creates a new consumer using the given broker addresses and configuration
	conn, err := sarama.NewConsumer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

var msgCount int

var x int64

func consume() {

	printMessage("Consumer started ")

	topic := "comments"
	worker, err := connectConsumer([]string{"kafka:9092"})
	if err != nil {
		fmt.Println("NO CONNECTION CONSUMER")
		panic(err)
	}

	// Calling ConsumePartition. It will open one connection per broker
	// and share it for all partitions that live on it.

	consumer, err := worker.ConsumePartition(topic, 0, x)

	defer consumer.Close()

	if err != nil {
		fmt.Println("TUTSI")
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

			db := Init()
			// dynamic
			insertDynStmt := `insert into "test"("comment") values($1)`

			_, err_db := db.Exec(insertDynStmt, cmt)

			checkErr(err_db)

			printMessage("Inserting comment into DB")

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

//...........................................
