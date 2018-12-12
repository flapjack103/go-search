package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// DefaultServerPort is the port the http server listens on
// XXX: allow this to be overridden
const DefaultServerPort = "8080"

// Server is the http server for handling queries and serving the static
// files for the website.
type Server struct {
	mux     *http.ServeMux
	port    string
	querier *Querier
	fileMgr *FileManager
}

// NewServer inits an http server and attaches the handlers.
func NewServer(q *Querier, m *FileManager) *Server {
	port := DefaultServerPort
	if envPort, ok := os.LookupEnv("PORT"); ok {
		port = envPort
	}

	s := &Server{
		mux:     http.NewServeMux(),
		port:    port,
		querier: q,
		fileMgr: m,
	}

	s.initHandlers()

	return s
}

// Listen starts the server listening for requests.
func (s *Server) Listen() error {
	fmt.Printf("Listening on port %s...\n", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), s.mux)
}

func (s *Server) initHandlers() {
	// search static assets
	fs := http.FileServer(http.Dir("static"))
	s.mux.Handle("/", fs)

	// endpoints for dynamically requesting data
	s.mux.HandleFunc("/summary.json", s.summaryHandler)
	s.mux.HandleFunc("/preview", s.previewHandler)
	s.mux.HandleFunc("/search", s.searchHandler)
}

/* Request Handler Functions */

func (s *Server) summaryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.String())

	summary := s.querier.idx.Summary()
	data, err := json.Marshal(summary)
	if err != nil {
		fmt.Printf("Error creating summary: %s\n", err)
		fmt.Fprint(w, "{\"error\": \"error generating summary\"}")
		return
	}
	fmt.Fprint(w, string(data))
}

func (s *Server) previewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.String())

	params := r.URL.Query()
	if len(params) == 0 {
		fmt.Fprint(w, "")
		return
	}

	file, ok := params["file"]
	if !ok {
		fmt.Fprint(w, "{\"error\": \"must specify file\"")
		return
	}

	line, ok := params["line"]
	if !ok {
		fmt.Fprint(w, "{\"error\": \"must specify line number\"}")
		return
	}

	if len(line) == 0 || len(file) == 0 {
		fmt.Fprint(w, "{\"error\": \"not enough arguments\"}")
		return
	}

	lineNum, err := strconv.Atoi(line[0])
	if err != nil {
		fmt.Fprint(w, "{\"error\": \"line must be an integer\"}")
		return
	}

	preview, err := s.fileMgr.GetFilePreview(file[0], lineNum)
	if err != nil {
		fmt.Printf("Error generating preview for file %s line %d: %s\n", file, lineNum, err)
		fmt.Fprint(w, "{\"error\": \"error generating preview\"}")
		return
	}

	data, err := json.Marshal(preview)
	if err != nil {
		fmt.Printf("Error marshalling preview response: %s\n", err)
		fmt.Fprint(w, "{\"error\": \"error generating preview\"}")
		return
	}
	fmt.Fprint(w, string(data))
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.String())

	params := r.URL.Query()
	if len(params) == 0 {
		fmt.Fprint(w, "")
		return
	}
	query, ok := params["query"]
	if !ok {
		fmt.Fprint(w, "")
		return
	}

	// parse possible query filters
	opts := DefaultQueryOptions()
	if file, ok := params["file"]; ok {
		opts.file = strings.ToLower(file[0])
	}
	if wtype, ok := params["type"]; ok {
		opts.wtype = strings.ToLower(wtype[0])
	}
	if limit, ok := params["limit"]; ok {
		if l, err := strconv.Atoi(limit[0]); err == nil {
			opts.limit = l
		}
	}

	refs := s.querier.Query(strings.Join(query, ""), opts)
	data, err := json.Marshal(refs.Format())
	if err != nil {
		fmt.Printf("Error running search: %s\n", err)
		fmt.Fprint(w, fmt.Sprintf("{\"error\":\"%s\"}", err))
		return
	}
	fmt.Fprint(w, string(data))
}
