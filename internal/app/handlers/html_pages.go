package handlers

import "net/http"

func Empty(w http.ResponseWriter, r *http.Request) {

	data := []byte("HELLO WORLD! I'm Ilia!\n Welcome! try /plotnikov!")
	w.WriteHeader(200)
	w.Write(data)
}

func PageDB(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "./html/db.html")
}
func PagePost(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "./html/post.html")

}
