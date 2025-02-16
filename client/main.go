package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

/*
O Client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar.
Precisará receber do server.go apenas o valor atual do câmbio (campo "bid" do JSON).
Utilizando o package "context", terá um timeout máximo de 300ms para receber o resultado
Terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
*/

const (
	timeout  = 300 * time.Millisecond
	URL      = "http://localhost:8080/cotacao"
	FILENAME = "cotacao.txt"
)

type Exchange struct {
	Value string `json:"value"`
}

func main() {
	ctx := context.Background()

	response, err := DoGetExchangeRequest(ctx)
	if err != nil {
		panic(err)
	}

	exchange, err := ParseExchange(response)
	if err != nil {
		panic(err)
	}

	err = WriteInFile(exchange)
	if err != nil {
		panic(err)
	}
}

// função que realiza a requisição HTTP para o server.go
func DoGetExchangeRequest(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return []byte{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer CloseResponseBody(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

// Função que faz o parse do JSON da resposta da requisição
func ParseExchange(dataJson []byte) (Exchange, error) {
	var exchange Exchange
	err := json.Unmarshal(dataJson, &exchange)
	if err != nil {
		return Exchange{}, err
	}
	return exchange, nil
}

// Função que escreve o valor da cotação em um arquivo
func WriteInFile(exchange Exchange) error {
	file, err := os.Create(FILENAME)
	if err != nil {
		return err
	}
	defer CloseFile(file)
	_, err = fmt.Fprintf(file, "Dólar: %s", exchange.Value)
	if err != nil {
		return err
	}
	return nil
}

// Função que fecha o corpo da resposta
func CloseResponseBody(body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		panic(err)
	}
}

// Função que fecha o arquivo
func CloseFile(file *os.File) {
	err := file.Close()
	if err != nil {
		panic(err)
	}
}
