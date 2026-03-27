package util

import (
	"encoding/json"
	"net/http"
)

func CheckPostReq(rw *http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(*rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func CheckGetReq(rw *http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(*rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func DecodeReq(req *http.Request, str any) {
	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(str)
	if err != nil {
		panic(err)
	}
}
