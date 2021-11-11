package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

// HSTSMaxAge is max-age of HSTS
const HSTSMaxAge = 6 * 30 * 24 * 3600

var Endpoint = os.Getenv("DYNAMODB_ENDPOINT")
var TableName = os.Getenv("DYNAMODB_TABLE_NAME")

type item struct {
	ID     string `json:"id" dynamo:"id"`
	Status int    `json:"status,omitempty" dynamo:"status"`
	To     string `json:"to" dynamo:"to"`
}

func (item item) toJSON() []byte {
	b, _ := json.Marshal(item) // never fails, so ignore error
	return b
}

func table() (*dynamo.Table, error) {
	cfg := aws.NewConfig()
	if Endpoint != "" {
		cfg = cfg.WithEndpoint(Endpoint)
	}
	s, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	db := dynamo.New(s, cfg)
	t := db.Table(TableName)
	return &t, nil
}

func get(key string) (*item, error) {
	t, err := table()
	if err != nil {
		return nil, err
	}
	var item item
	err = t.Get("id", key).One(&item)
	return &item, err
}

func addHSTS(w http.ResponseWriter) {
	w.Header().Add("Strict-Transport-Security", fmt.Sprintf("max-age=%d", HSTSMaxAge))
}

func redirect(w http.ResponseWriter, location string, status int) {
	if !(300 <= status && status < 400) {
		status = http.StatusTemporaryRedirect
	}
	addHSTS(w)
	http.Redirect(w, &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/"}}, location, status)
}

func returnIndexError(w http.ResponseWriter, err error, key string) {
	w.Header().Add("Content-Type", "text/html")
	if err == dynamo.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<!doctype html><title>no such key</title>\n"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<!doctype html><title>something went wrong</title>\n"))
	}
	w.Write([]byte(fmt.Sprintf("err: %v<br />\n", err.Error())))
	if i := strings.LastIndex(key, "/"); i > 0 { // keyはURIのpathなので、必ず/から始まるが、それを除いて/を含んでいるもの。
		prefixKey := key[:i]
		w.Write([]byte(fmt.Sprintf("maybe here?: <a href='%v'>%v</a>", prefixKey, prefixKey))) // そこへのリンクを出しておく。
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	addHSTS(w)
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
		returnIndexError(w, err, key)
		return
	}
	if expand {
		w.Header().Add("Content-Type", "application/json")
		w.Write(item.toJSON())
		return
	}
	redirect(w, item.To, item.Status)
}

func main() {
	http.HandleFunc("/", indexHandler)
	algnhsa.ListenAndServe(nil, &algnhsa.Options{RequestType: algnhsa.RequestTypeAPIGateway})
}
