package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultServerPort = 8080

type Server struct {
	mux     *http.ServeMux
	port    int
	querier *Querier
}

func NewServer(q *Querier) *Server {
	s := &Server{
		mux:     http.NewServeMux(),
		port:    defaultServerPort,
		querier: q,
	}

	s.initHandlers()

	return s
}

func (s *Server) Listen() error {
	fmt.Printf("Listening on port %d...\n", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux)
}

func (s *Server) initHandlers() {
	// search static assets
	fs := http.FileServer(http.Dir("static"))
	s.mux.Handle("/", fs)

	// endpoints for dynamically requesting data
	s.mux.HandleFunc("/functions.json", s.functionsHandler)
	s.mux.HandleFunc("/summary.json", s.summaryHandler)
	s.mux.HandleFunc("/search", s.searchHandler)
}

func (s *Server) functionsHandler(w http.ResponseWriter, r *http.Request) {
	refs := s.querier.Query("")
	data, err := json.Marshal(refs.Format())
	if err != nil {
		fmt.Printf("Error running search: %s\n", err)
		fmt.Fprint(w, fmt.Sprintf("{error:%s}", err))
		return
	}
	fmt.Fprint(w, string(data))
}

func (s *Server) summaryHandler(w http.ResponseWriter, r *http.Request) {
	summary := s.querier.idx.Summary()
	data, err := json.Marshal(summary)
	if err != nil {
		fmt.Printf("Error creating summary: %s\n", err)
		fmt.Fprint(w, fmt.Sprintf("{error:%s}", err))
		return
	}
	fmt.Fprint(w, string(data))
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if len(params) == 0 {
		fmt.Fprint(w, "")
	}
	query, ok := params["query"]
	if !ok {
		fmt.Fprint(w, "")
	}

	refs := s.querier.Query(strings.Join(query, ""))
	data, err := json.Marshal(refs.Format())
	if err != nil {
		fmt.Printf("Error running search: %s\n", err)
		fmt.Fprint(w, fmt.Sprintf("{error:%s}", err))
		return
	}
	fmt.Fprint(w, string(data))
}
