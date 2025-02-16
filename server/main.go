package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func CotacaoHandler(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write([]byte("{\"value\": \"5.695\"}"))
	if err != nil {
		panic(err)
	}
}
