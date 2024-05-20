package api

import (
	"encoding/json"
	"fmt"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/config"
	"github.com/zooper-corp/CoinWatch/data"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ApiServer struct {
	config config.ApiServerConfig
	client *client.Client
}

type ApiResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewApiServer(c *client.Client, cfg config.ApiServerConfig) ApiServer {
	return ApiServer{cfg, c}
}

func (s *ApiServer) Start() {
	http.HandleFunc("/api/v1/balance", s.corsMiddleware(s.authMiddleware(s.handleBalance)))
	http.HandleFunc("/api/v1/history", s.corsMiddleware(s.authMiddleware(s.handleHistory)))
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("Starting API server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Unable to start API server: %v", err)
	}
}

func (s *ApiServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		apiKey := r.Header.Get("X-Api-Key")
		if apiKey != s.config.ApiKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (s *ApiServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (s *ApiServer) handleBalance(w http.ResponseWriter, r *http.Request) {
	balance := s.client.GetLastBalance()
	response := ApiResponse{
		Message: "Balance retrieved successfully",
		Data:    balance.Entries(),
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (s *ApiServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	amountStr := r.URL.Query().Get("amount")
	intervalStr := r.URL.Query().Get("interval")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid amount parameter", http.StatusBadRequest)
		return
	}
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval <= 0 {
		http.Error(w, "Invalid interval parameter", http.StatusBadRequest)
		return
	}

	totalHours := amount * interval
	totalDays := (totalHours + 23) / 24 // ceil division to ensure full coverage
	bs, err := s.client.QueryBalance(data.BalanceQueryOptions{Days: totalDays})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to query balances: %v", err), http.StatusInternalServerError)
		return
	}

	// Get tokens
	unsortedTokens := bs.Tokens()
	entries := bs.GetTimeSeries(amount, time.Duration(interval)*time.Hour)
	lastSampleTotals := make([]float64, len(unsortedTokens))
	for i, t := range unsortedTokens {
		lastSampleTotals[i] = entries[0].FilterToken(t).TotalFiatValue()
	}
	order := tools.ReverseIntArray(tools.SortAndReturnIndex(lastSampleTotals))
	tokens := make([]string, len(unsortedTokens))
	for i, idx := range order {
		tokens[i] = unsortedTokens[idx]
	}

	// Create a series for every token axing at max items
	dataSeries := make([]map[string]interface{}, 0)
	currentTime := time.Now()
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry.Entries()) > 0 {
			point := make(map[string]interface{})
			point["timestamp"] = currentTime.Add(time.Duration(-i*interval) * time.Hour)
			balances := make([]map[string]interface{}, 0)
			for _, token := range tokens {
				tokenData := entry.FilterToken(token)
				balances = append(balances, map[string]interface{}{
					"token":      strings.ToLower(token),
					"balance":    tokenData.TokenBalance(token),
					"fiat_value": tokenData.TotalFiatValue(),
				})
			}
			point["balance"] = balances
			dataSeries = append(dataSeries, point)
		}
	}

	response := ApiResponse{
		Message: "History retrieved successfully",
		Data:    dataSeries,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
