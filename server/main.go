package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CotacaoJson struct {
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

type Cotacao struct {
	ID         int `gorm:"primaryKey"`
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

var client = &http.Client{Timeout: 5 * time.Second}

// var client = &http.Client{Timeout: 200 * time.Millisecond}
func main() {
	http.HandleFunc("/cotacao", HandleRequest)
	log.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	log.Println("Request initiated")

	apiCtx, apiCancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer apiCancel()

	cotacaoJson, err := findCotacaoFromApi(apiCtx)
	if err != nil {
		log.Println("Failed to fetch quotation data:", err)
		http.Error(w, "Failed to fetch quotation data", http.StatusInternalServerError)
		return
	}

	cotacao := Cotacao{
		Code:       cotacaoJson.USDBRL.Code,
		Codein:     cotacaoJson.USDBRL.Codein,
		Name:       cotacaoJson.USDBRL.Name,
		High:       cotacaoJson.USDBRL.High,
		Low:        cotacaoJson.USDBRL.Low,
		VarBid:     cotacaoJson.USDBRL.VarBid,
		PctChange:  cotacaoJson.USDBRL.PctChange,
		Bid:        cotacaoJson.USDBRL.Bid,
		Ask:        cotacaoJson.USDBRL.Ask,
		Timestamp:  cotacaoJson.USDBRL.Timestamp,
		CreateDate: cotacaoJson.USDBRL.CreateDate,
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbCancel()

	if err := saveCotacaoToDb(dbCtx, cotacao); err != nil {
		log.Println("Failed to save quotation to DB:", err)
		http.Error(w, "Failed to save quotation to DB", http.StatusInternalServerError)
		return
	}

	log.Println("Quotation created successfully:", cotacao)
	w.Write([]byte(cotacao.Bid))
}

func findCotacaoFromApi(ctx context.Context) (CotacaoJson, error) {
	var cotacao CotacaoJson
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return cotacao, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return cotacao, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cotacao, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		return cotacao, fmt.Errorf("failed to decode response body: %w", err)
	}

	return cotacao, nil
}

func saveCotacaoToDb(ctx context.Context, cotacao Cotacao) error {
	db, err := gorm.Open(sqlite.Open("./data/database.db"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(&Cotacao{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := db.WithContext(ctx).Create(&cotacao).Error; err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	return nil
}
