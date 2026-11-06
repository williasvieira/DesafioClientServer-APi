package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRateResponse struct {
	USDBRL struct {
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

type Server struct {
	db *sql.DB
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatalf("Erro ao abrir banco de dados SQLite: %v", err)
	}
	defer db.Close()

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Erro ao criar tabela no banco de dados: %v", err)
	}

	server := &Server{db: db}

	http.HandleFunc("/cotacao", server.handleCotacao)
	log.Println("Servidor rodando na porta :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func (s *Server) handleCotacao(w http.ResponseWriter, r *http.Request) {
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancelAPI()

	cotacao, err := fetchExchangeRate(ctxAPI)
	if err != nil {
		log.Printf("[SERVER ERROR] Falha ao buscar cotação na API externa: %v", err)
		http.Error(w, "Erro ao buscar cotação externa (timeout/erro)", http.StatusGatewayTimeout)
		return
	}

	ctxDB, cancelDB := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancelDB()

	err = s.saveCotacao(ctxDB, cotacao.USDBRL.Bid)
	if err != nil {
		log.Printf("[SERVER ERROR] Falha ao salvar cotação no banco de dados: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao)
}

func fetchExchangeRate(ctx context.Context) (*ExchangeRateResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status inesperado da API: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data ExchangeRateResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *Server) saveCotacao(ctx context.Context, bid string) error {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, bid)
	if err != nil {
		return err
	}

	return nil
}
