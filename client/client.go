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

type Cotacao struct {
	Usdbrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, error := io.ReadAll(response.Body)
	if error != nil {
		panic(err)
	}
	var data Cotacao
	err = json.Unmarshal(body, &data)
	if error != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", data)

	salvarCotacaoEmArquivo(data)
}

func salvarCotacaoEmArquivo(cotacao Cotacao) {
	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString(fmt.Sprintf("Dólar: %s\n", cotacao.Usdbrl.Bid))
}
