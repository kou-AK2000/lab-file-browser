package server

import "net/http"

func NewServer() *http.Server {

	mux := http.NewServeMux()
	registerRoutes(mux)

	return &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}
}
