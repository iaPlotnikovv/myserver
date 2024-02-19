package endpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchInfo(w http.ResponseWriter, r *http.Request) {

	zapis := r.FormValue("comment")

	if zapis != "" {

		response, err := Caching(zapis)

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

		PrintMessage("Getting info...")

	} else {

		ShowDB(w, r)

	}
}
