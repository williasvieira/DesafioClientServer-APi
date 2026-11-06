package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	// Timeout de 300ms para a requisição ao servidor local
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("[CLIENT ERROR] Erro ao criar requisição: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[CLIENT ERROR] Timeout ou erro ao consultar servidor: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[CLIENT ERROR] Resposta com código HTTP inválido: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[CLIENT ERROR] Erro ao ler resposta do servidor: %v", err)
		return
	}

	var cotacao CotacaoResponse
	if err := json.Unmarshal(body, &cotacao); err != nil {
		log.Printf("[CLIENT ERROR] Erro ao fazer parse do JSON: %v", err)
		return
	}

	if cotacao.USDBRL.Bid == "" {
		log.Println("[CLIENT ERROR] Cotação em branco recebida do servidor.")
		return
	}

	// Grava o arquivo cotacao.txt
	content := fmt.Sprintf("Dólar: %s", cotacao.USDBRL.Bid)
	err = os.WriteFile("cotacao.txt", []byte(content), 0644)
	if err != nil {
		log.Fatalf("[CLIENT ERROR] Erro ao salvar arquivo cotacao.txt: %v", err)
	}

	log.Printf("Sucesso! Cotação salva em cotacao.txt: %s\n", content)
}
