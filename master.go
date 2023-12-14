package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "test"
	dbname   = "mydb"
)

// инициализируем соединение с БД
var db *sql.DB

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
		fmt.Printf("Ошибка ???, %s", err)
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

	//mux.HandleFunc("/plotnikov/db", PostInfo).Methods("POST")

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
	Type    string    `json:"type"`
	Data    []info_js `json:"data"`
	Message string    `json:"message"`
}

func GetInfo(res http.ResponseWriter, req *http.Request) {

	//db := Init()

	printMessage("Getting info...")

	// Get all movies from movies table that don't have comment = "1"
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
	// loop and display the result in the browser
	for _, snb := range info {
		fmt.Fprintf(res, "%d %s\n", snb.ID, snb.Comment)
	}

	var response = JsonResponse{Type: "success", Data: info}

	json.NewEncoder(res).Encode(response)
}
