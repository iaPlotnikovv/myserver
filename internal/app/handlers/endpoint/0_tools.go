package endpoint

import "fmt"

func PrintMessage(message string) {
	fmt.Println("")
	fmt.Println(message)
	fmt.Println("")
}

func CheckErr(err error) {
	if err != nil {
		fmt.Printf("Ошибка, %s", err)
		panic(err)
	}
}

type info_js struct {
	ID      int    `json:"id"`
	Comment string `json:"comment"`
}

type JsonResponse struct {
	Type string    `json:"type"`
	Data []info_js `json:"data"`
}
