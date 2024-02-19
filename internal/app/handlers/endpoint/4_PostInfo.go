package endpoint

import (
	"encoding/json"
	"net/http"
)

func PostInfo(w http.ResponseWriter, r *http.Request) {

	comment := r.FormValue("comment")
	mesg := []byte(comment)

	var response = JsonResponse{}

	if comment != "" {

		PushCommentToQueue("comments", mesg)

		Consume()

	} else {
		response = JsonResponse{Type: "error"}
	}

	json.NewEncoder(w).Encode(response)
}
