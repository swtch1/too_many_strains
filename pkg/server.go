package tms

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	// Port is the port where the server will listen.
	Port int32
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	// TODO: ensure only the GET verb can be called on this endpoint
	//mux.HandleFunc("/", s.placeholder)
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return httpSrv.ListenAndServe()
}
