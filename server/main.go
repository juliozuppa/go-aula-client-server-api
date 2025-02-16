package main

import (
	"context"
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"time"
)

/*
O server.go deverá consumir a API contendo o câmbio de Dólar e Real
no endereço: https://economia.awesomeapi.com.br/json/last/USD-BRL e em seguida
deverá retornar no formato JSON o resultado para o cliente.
Usando o package "context", o server.go deverá registrar no banco de dados SQLite
cada cotação recebida, sendo que o timeout máximo para chamar a API de cotação do
dólar deverá ser de 200ms e o timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms.
*/
const (
	ApiUrl     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	ApiTimeout = 200 * time.Millisecond
	DbTimeout  = 10 * time.Millisecond
	DbFile     = "exchange.db"
)

type Exchange struct {
	ID         int    `gorm:"primaryKey" json:"-"`
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Response struct {
	Exchange Exchange `json:"USDBRL"`
}

type ExchangeResponse struct {
	Value string `json:"value"`
}

func main() {
	db, err := InitDatabase()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", ExchangeHandler(db))
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// InitDatabase Função que inicializa o banco de dados
func InitDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(DbFile), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Exchange{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ExchangeHandler Função que trata a requisição HTTP
func ExchangeHandler(db *gorm.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		// consultar o valor atual do dolar
		ctx := context.Background()
		log.Println("Realizando consulta da cotação do dólar")
		response, err := DoGetExchangeRequest(ctx)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Tratando a resposta da consulta")
		exchange, err := ParseExchange(response)
		if err != nil {
			log.Fatal(err)
		}

		// registrar no banco de dados a cotação
		log.Println("Registrando a cotação no banco de dados")
		WriteInDB(ctx, db, exchange.Exchange)

		err = SendResponse(writer, exchange.Exchange)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// WriteInDB Função que registra a cotação no banco de dados
func WriteInDB(ctx context.Context, db *gorm.DB, exchange Exchange) {
	ctx, cancel := context.WithTimeout(ctx, DbTimeout)
	defer cancel()
	db.WithContext(ctx).Create(&exchange)
}

// SendResponse Função que envia a resposta ao cliente
func SendResponse(writer http.ResponseWriter, exchange Exchange) error {
	response := ExchangeResponse{Value: exchange.Bid}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	log.Println("Respondendo ao cliente", response)
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

// DoGetExchangeRequest função que realiza a requisição HTTP para o serviço externo
func DoGetExchangeRequest(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, ApiTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ApiUrl, nil)
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

// ParseExchange Função que faz o parse do JSON da resposta da requisição
func ParseExchange(dataJson []byte) (Response, error) {
	var response Response
	err := json.Unmarshal(dataJson, &response)
	if err != nil {
		return Response{}, err
	}
	return response, nil
}

// CloseResponseBody Função que fecha o corpo da resposta
func CloseResponseBody(body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		log.Fatal(err)
	}
}
