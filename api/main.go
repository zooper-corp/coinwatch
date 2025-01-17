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
	Updated time.Time   `json:"updated,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewApiServer(c *client.Client, cfg config.ApiServerConfig) ApiServer {
	return ApiServer{cfg, c}
}

func (s *ApiServer) Start() {
	http.HandleFunc("/api/v1/balance", s.corsMiddleware(s.authMiddleware(s.handleBalance)))
	http.HandleFunc("/api/v1/history", s.corsMiddleware(s.authMiddleware(s.handleHistory)))
	http.HandleFunc("/api/v1/query", s.corsMiddleware(s.authMiddleware(s.handleQuery)))
	http.HandleFunc("/metrics", s.corsMiddleware(s.authMiddleware(s.handleMetrics)))
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

func (s *ApiServer) writeJSONResponse(w http.ResponseWriter, response ApiResponse) {
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(s.config.CacheTTL.Seconds())))
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (s *ApiServer) handleBalance(w http.ResponseWriter, r *http.Request) {
	balance := s.client.GetLastBalance()
	response := ApiResponse{
		Message: "Balance retrieved successfully",
		Updated: s.client.GetLastBalanceUpdate(),
		Data:    balance.Entries(),
	}
	s.writeJSONResponse(w, response)
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
		Updated: s.client.GetLastBalanceUpdate(),
		Data:    dataSeries,
	}
	s.writeJSONResponse(w, response)
}

func (s *ApiServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	balance := s.client.GetLastBalance()
	// Set headers
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(s.config.CacheTTL.Seconds())))
	w.Header().Set("Content-Type", "text/plain")
	// Find BTC value first
	btcValue := balance.FilterToken("BTC").TotalFiatValue()
	btcPrice := btcValue / balance.FilterToken("BTC").TokenBalance("BTC")
	// Write metrics
	tokens := balance.Tokens()
	for _, token := range tokens {
		bd := balance.FilterToken(token)
		amount := bd.TokenBalance(token)
		value := bd.TotalFiatValue()
		price := value / amount
		metrics := []string{
			fmt.Sprintf("crypto_balance{token=\"%s\",type=\"balance\"} %f\n", token, amount),
			fmt.Sprintf("crypto_balance{token=\"%s\",type=\"value\"} %f\n", token, value),
			fmt.Sprintf("crypto_balance{token=\"%s\",type=\"price\"} %f\n", token, price),
			fmt.Sprintf("crypto_balance{token=\"%s\",type=\"price_btc\"} %f\n", token, price/btcPrice),
			fmt.Sprintf("crypto_balance{token=\"%s\",type=\"value_btc\"} %f\n", token, value/btcPrice),
		}
		for _, metric := range metrics {
			if _, err := w.Write([]byte(metric)); err != nil {
				log.Printf("Error writing metric for token %s: %v", token, err)
				return
			}
		}
	}
}

func (s *ApiServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	intervalStr := r.URL.Query().Get("interval")
	mode := r.URL.Query().Get("mode")
	// Validate parameters
	if mode != "fiat_value" && mode != "token" && mode != "price" {
		http.Error(w, "Invalid 'mode' parameter. Allowed values: 'token', 'fiat_value' or 'price'", http.StatusBadRequest)
		return
	}
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}
	intervalHours, err := strconv.Atoi(intervalStr)
	if err != nil || intervalHours <= 0 {
		http.Error(w, "Invalid 'interval' parameter", http.StatusBadRequest)
		return
	}
	interval := time.Duration(intervalHours) * time.Hour
	// Fetch raw balances for [from, to]
	rawBalances, err := s.client.GetBalancesFromDate(from)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch balances: %v", err), http.StatusInternalServerError)
		return
	}
	// Build a time-series via GetTimeSeries.
	to := time.Now()
	totalDuration := to.Sub(from)
	if totalDuration < 0 {
		http.Error(w, "'from' must be before 'to'", http.StatusBadRequest)
		return
	}
	steps := int(totalDuration / interval)
	if steps < 1 {
		steps = 1
	}
	entries := rawBalances.GetTimeSeries(steps, interval)
	// Get tokens
	unsortedTokens := rawBalances.Tokens()
	lastSampleTotals := make([]float64, len(unsortedTokens))
	for i, t := range unsortedTokens {
		lastSampleTotals[i] = entries[0].FilterToken(t).TotalFiatValue()
	}
	order := tools.ReverseIntArray(tools.SortAndReturnIndex(lastSampleTotals))
	tokens := make([]string, len(unsortedTokens))
	for i, idx := range order {
		tokens[i] = unsortedTokens[idx]
	}
	// Prepare the final result
	result := make([]map[string]interface{}, 0)
	currentTime := time.Now()
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry.Entries()) > 0 {
			point := make(map[string]interface{})
			point["timestamp"] = currentTime.Add(time.Duration(-i*intervalHours) * time.Hour)
			for _, token := range tokens {
				tokenData := entry.FilterToken(token)
				if mode == "fiat_value" {
					point[token] = tokenData.TotalFiatValue()
				} else if mode == "token" {
					point[token] = tokenData.TokenBalance(token)
				} else if //goland:noinspection GoDfaConstantCondition
				mode == "price" {
					if tokenData.TokenBalance(token) > 0 {
						point[token] = tokenData.TotalFiatValue() / tokenData.TokenBalance(token)
					} else {
						continue
					}
				}
			}
			result = append(result, point)
		}
	}
	response := ApiResponse{
		Message: "Structured balances retrieved successfully",
		Data:    result,
	}
	s.writeJSONResponse(w, response)
}
