package main

import (
	"encoding/json"
	"net/http"
)

func fixCSR(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(501)
		return
	}
	w.Write(bs)
}

func showList(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)
		writeJSON(w, m.list())
	}
}

func UIServer(addr string, m *Manager) {
	http.Handle("/list", showList(m))

	http.ListenAndServe(addr, nil)
}
