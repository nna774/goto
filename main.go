package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

var endpoint = os.Getenv("DYNAMODB_ENDPOINT")
var table = os.Getenv("DYNAMODB_TABLE_NAME")

type item struct {
	ID     string `json:"id" dynamo:"id"`
	Status int    `json:"status,omitempty" dynamo:"status"`
	To     string `json:"to" dynamo:"to"`
}

func get(key string) (item, error) {
	cfg := aws.NewConfig()
	if endpoint != "" {
		cfg = cfg.WithEndpoint(endpoint)
	}
	db := dynamo.New(session.New(), cfg)
	table := db.Table(table)
	var item item
	err := table.Get("id", key).One(&item)
	return item, err
}

func redirect(w http.ResponseWriter, r *http.Request, location string, status int) {
	if !(300 <= status && status < 400) {
		status = http.StatusTemporaryRedirect
	}
	http.Redirect(w, r, location, status)
}

func write(w http.ResponseWriter, msg string) {
	w.Write([]byte(msg))
}

func (item item) toJSON() []byte {
	b, _ := json.Marshal(item)
	return b
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Write([]byte("helo"))
		return
	}
	key := r.URL.Path
	expand := false
	if strings.HasSuffix(key, "+") {
		expand = true
		key = key[:len(key)-1]
	}
	item, err := get(key)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if expand {
		w.Header().Add("Content-Type", "application/json")
		w.Write(item.toJSON())
		return
	}
	redirect(w, r, item.To, item.Status)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%v", nil)))
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/!add", addHandler)
	http.HandleFunc("/!add/", addHandler)
	algnhsa.ListenAndServe(http.DefaultServeMux, nil)
}
