package api

import (
	"encoding/json"
	"fmt"
	"github.com/zooper-corp/CoinWatch/client"
	"github.com/zooper-corp/CoinWatch/config"
	"log"
	"net/http"
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
	http.HandleFunc("/api/v1/balance", s.authMiddleware(s.handleBalance))
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
