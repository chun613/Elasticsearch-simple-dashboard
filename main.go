package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type ESData struct {
	Indices []Index
}

type Index struct {
	Health   string `json:"health"`
	Status   string `json:"status"`
	Index    string `json:"index"`
	UUID     string `json:"uuid"`
	DocCount string `json:"docs.count"`
	Size     string `json:"store.size"`
	Date     string `json:"creation.date.string"`
	Alias    string `json:"alias"`
}

type Alias struct {
	Alias string `json:"alias"`
	Index string `json:"index"`
}

var (
	tr              = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client          = &http.Client{Transport: tr}
	host            = "https://%s/api/v1/namespaces/high-memory-ha/services/elasticsearch-product-logging/proxy/"
	devHost         = "api.dev.zwoop.xyz"
	devUsername     = "admin"
	devPwd          = "0fTQE8F1LtzrZYe6SFlDwQTYReRnkHAi"
	london2Host     = "api.london-2.zwoop.xyz"
	london2Username = "admin"
	london2Pwd      = "KKfpTCY26Nlc6GjS3Fw834WBIFJYfhny"
	devDubHost      = "api.dev-dub.zwoop.xyz"
	devDubUsername  = "admin"
	devDubPwd       = "rW8jSQfAVVCfyKocII9eJEhx5kJMP7Td"
)

func main() {
	fmt.Println("let's get started")
	r := mux.NewRouter()
	r.HandleFunc("/{env}/indices", catIndicesHandler)
	r.HandleFunc("/{env}/templates", catTemplatesHandler)
	r.HandleFunc("/{env}/mapping/{index}", getMappingHandler)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func catIndicesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	// request for alias
	aliasesJSON := fireRequest("_cat/aliases?v&h=alias,index&format=json", r)
	var aliasData []Alias
	if err := json.Unmarshal([]byte(aliasesJSON), &aliasData); err != nil {
		panic(err)
	}
	// build to map for fast lookup
	var aliasMap = make(map[string]string)
	for _, v := range aliasData {
		aliasMap[v.Index] = v.Alias
	}

	// request for index
	indicesJSON := fireRequest("_cat/indices?v&h=health,status,index,uuid,docs.count,store.size,creation.date.string&format=json", r)
	var indicesData []Index
	if err := json.Unmarshal([]byte(indicesJSON), &indicesData); err != nil {
		panic(err)
	}

	// set alias
	for i, v := range indicesData {
		if alias, ok := aliasMap[v.Index]; ok {
			indicesData[i].Alias = alias
		}
	}

	j, _ := json.Marshal(indicesData)
	fmt.Fprintf(w, string(j))
}

func catTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	templateJSON := fireRequest("_cat/templates?v&format=json", r)
	fmt.Fprintf(w, templateJSON)
}

func getMappingHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	vars := mux.Vars(r)
	index := vars["index"]
	mappingJSON := fireRequest(index, r)
	fmt.Fprintf(w, mappingJSON)
}

func getEnvVar(r *http.Request) (string, string, string) {
	vars := mux.Vars(r)
	env := vars["env"]
	if env == "dev" {
		return fmt.Sprintf(host, devHost), devUsername, devPwd
	} else if env == "london-2" {
		return fmt.Sprintf(host, london2Host), london2Username, london2Pwd
	} else if env == "dev-dub" {
		return fmt.Sprintf(host, devDubHost), devDubUsername, devDubPwd
	}
	panic("unknown env")
}

func fireRequest(endpoint string, r *http.Request) string {
	host, username, pwd := getEnvVar(r)

	req, _ := http.NewRequest("GET", host+endpoint, nil)
	req.SetBasicAuth(username, pwd)
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	data, _ := ioutil.ReadAll(resp.Body)
	return string(data)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
