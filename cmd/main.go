package main

import (
	"net/http"

	h "github.com/iaPlotnikovv/myserver/internal/app/handlers"
	e "github.com/iaPlotnikovv/myserver/internal/app/handlers/endpoint"
)

func main() {

	//create mux

	mux := http.NewServeMux()

	mux.HandleFunc("/", h.Empty)

	mux.HandleFunc("/plotnikov", h.PageDB)

	mux.HandleFunc("/plotnikov/db_post", h.PagePost)

	mux.HandleFunc("/plotnikov/db", e.FetchInfo)

	mux.HandleFunc("/plotnikov/db_post/", e.PostInfo)

	//server

	s := &http.Server{
		Addr:    ":1311",
		Handler: mux,
	}

	s.ListenAndServe()

}
