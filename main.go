package main

import (
	"net/http"

	"github.com/akrylysov/algnhsa"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.URL.Path))
}

func main() {
	http.HandleFunc("/", indexHandler)
	algnhsa.ListenAndServe(http.DefaultServeMux, nil)
}
