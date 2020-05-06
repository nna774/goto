package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/nna774/lambda-authkun/adapter"
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

func list() ([]item, error) {
	cfg := aws.NewConfig()
	if endpoint != "" {
		cfg = cfg.WithEndpoint(endpoint)
	}
	db := dynamo.New(session.New(), cfg)
	table := db.Table(table)
	var items []item
	err := table.Scan().All(&items)
	return items, err
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

func startAuthHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	if err := t.Execute(w, nil); err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	items, err := list()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	res, err := json.Marshal(items)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(res)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/_", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/_/", http.StatusFound) })
	http.HandleFunc("/_/", startAuthHandler) // もっといい区切り文字使いたかったけど、API Gatewayの制限であんまり選べなかった。
	http.HandleFunc("/_/list", listHandler)
	http.HandleFunc("/_auth/callback", adapter.NewCallbackHandler("https://auth.dark-kuins.net/callback"))
	algnhsa.ListenAndServe(nil, &algnhsa.Options{RequestType: algnhsa.RequestTypeAPIGateway})
}
