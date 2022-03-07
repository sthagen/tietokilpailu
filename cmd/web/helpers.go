package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func (s *server) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	ts, ok := s.templateCache[name]
	if !ok {
		s.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, s.addDefaultData(td, r))
	if err != nil {
		s.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

func (s *server) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	return td
}
