package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type addParam struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type addResult struct {
	Code int `json:"code"`
	Data int `json:"data"`
}

func add(x, y int) int {
	return x + y
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	var param addParam
	json.Unmarshal(b, &param)

	ret := add(param.X, param.Y)

	respBytes, _ := json.Marshal(addResult{Code: 0, Data: ret})
	w.Write(respBytes)
}

func main() {
	http.HandleFunc("/add", addHandler)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
