package endpoint

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iaPlotnikovv/myserver/internal/app/init/database"
	_ "github.com/lib/pq"
)

var rows *sql.Rows
var err error

func ShowDB(w http.ResponseWriter, r *http.Request) {

	db := database.Init()

	rows, err = db.Query("SELECT * FROM test")

	CheckErr(err)

	PrintMessage("This is DataBase...")

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
