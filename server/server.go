package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	db *sql.DB
}

type CotacaoAPIResponse struct {
	Usdbrl struct {
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
	} `json:"USDBRL"`
}

func main() {
	// abri conexao com a base de dados
	db, err := sql.Open("sqlite3", "../data.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	server := &Server{db: db}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, cotacao_response TEXT NOT NULL)")
	if err != nil {
		panic(err)
	}
	// sobe servidor http
	http.HandleFunc("/cotacao", server.GetCotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func (server *Server) GetCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctxCotacaoRequest, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ctxCotacaoDBInsert, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	cotacaoResponse, err := GetCotacao(ctxCotacaoRequest)

	if err != nil {
		fmt.Println("error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = SalvaCotacao(ctxCotacaoDBInsert, server.db, cotacaoResponse)
	if err != nil {
		fmt.Println("error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacaoResponse)
}

func GetCotacao(ctx context.Context) (*CotacaoAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, error := ioutil.ReadAll(res.Body)
	if error != nil {
		return nil, err
	}
	var data CotacaoAPIResponse
	err = json.Unmarshal(body, &data)
	if error != nil {
		return nil, err
	}

	return &data, nil
}

func SalvaCotacao(ctx context.Context, db *sql.DB, cotacaoResponse *CotacaoAPIResponse) error {
	stmt, err := db.Prepare("insert into cotacoes (cotacao_response) values (?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	json, err := json.Marshal(cotacaoResponse)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, json)
	if err != nil {
		return err
	}

	return nil
}
