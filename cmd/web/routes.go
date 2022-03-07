package main

import (
	"net/http"
)

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.defaultHandler)
	mux.HandleFunc("/go/ping", ping)
	handler := s.quizHandler()
	mux.HandleFunc(s.baseUrl, handler)
	fs := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))

	return secureHeaders(mux)
}
