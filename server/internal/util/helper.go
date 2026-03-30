package util

import (
	"encoding/json"
	"net/http"
)

func DecodeReq(req *http.Request, str any) {
	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(str)
	if err != nil {
		panic(err)
	}
}
