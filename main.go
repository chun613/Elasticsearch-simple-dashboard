package main

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

type ESData struct {
	Indices string
}

var (
	tr       = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client   = &http.Client{Transport: tr}
	host     = "https://xxxx/api/v1/namespaces/high-memory-ha/services/elasticsearch-product-logging/proxy/"
	username = ""
	pwd      = ""
)

func main() {
	fmt.Println("let's get started")
	indices := fireRequest("_cat/indices?v")
	fmt.Println(indices)
	aliases := fireRequest("_cat/aliases?v")
	fmt.Println(aliases)

	tmpl := template.Must(template.ParseFiles("./template/index.html"))
	data := &ESData{Indices: indices}
	requestHandler(tmpl, data)
	http.ListenAndServe(":8080", nil)
}

func requestHandler(tmpl *template.Template, data *ESData) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	})
}

func fireRequest(endpoint string) string {
	req, _ := http.NewRequest("GET", host+endpoint, nil)
	setupHTTPRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	} else {
		data, _ := ioutil.ReadAll(resp.Body)
		return string(data)
	}
}

func setupHTTPRequest(req *http.Request) {
	req.SetBasicAuth(username, pwd)
	req.Header.Set("Content-type", "application/json")
}
